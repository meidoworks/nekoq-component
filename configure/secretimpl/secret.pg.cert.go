package secretimpl

import (
	"context"
	"crypto/x509"
	"database/sql"
	"errors"
	"time"

	"github.com/meidoworks/nekoq-component/configure/secretapi"
)

var _ secretapi.CertStorage = new(PostgresKeyStorage)

func (p *PostgresKeyStorage) saveInternal(certLevelType secretapi.CertLevelType, certName string, caCertSN secretapi.CertSerialNumber, cert *x509.Certificate, keyInfo secretapi.CertKeyInfo) (sn secretapi.CertSerialNumber, rerr error) {
	certData, err := new(secretapi.PemTool).EncodeCertificate(cert)
	if err != nil {
		return "", err
	}
	// root ca should not have signing cert
	if certLevelType == secretapi.CertLevelTypeRootCA {
		caCertSN = ""
	}

	tx, err := p.db.BeginTx(context.Background(), nil)
	if err != nil {
		return "", err
	}
	success := false
	defer func(tx *sql.Tx) {
		if !success {
			err := tx.Rollback()
			if err != nil {
				rerr = err
			}
		} else {
			err := tx.Commit()
			if err != nil {
				rerr = err
			}
		}
	}(tx)

	// query existing max version
	f := func() (int64, error) {
		var maxVersion *int64
		rows := tx.QueryRow("select max(cert_version) from secret_cert where cert_type = $1 and cert_name = $2", certLevelType, certName)
		if err := rows.Scan(&maxVersion); maxVersion == nil {
			return 0, nil
		} else if err != nil {
			return 0, err
		}
		return *maxVersion, nil
	}
	maxVersion, err := f()
	if err != nil {
		return "", err
	}
	nextVersion := maxVersion + 1

	now := time.Now()
	// mark existing certs archived
	r, err := tx.Exec(`
update secret_cert
set cert_status = 1, time_updated = $1
where cert_type = $2 and cert_name = $3;
`, now.UnixMilli(), certLevelType, certName)
	if err != nil {
		return "", err
	}
	if _, err := r.RowsAffected(); err != nil {
		return "", err
	}

	var certSN secretapi.CertSerialNumber
	certSN.FromBigInt(cert.SerialNumber)
	now = time.Now()
	// insert new cert
	r, err = tx.Exec(`
insert into secret_cert (cert_id, cert_type, cert_name, cert_version, cert_status, parent_ca_cert_id, cert_key_level,
                         cert_key_name, cert_content, expired_time, time_created, time_updated)
values ($1, $2, $3, $4, 0, $5, $6, $7, $8, $9, $10, $11);
`, certSN, certLevelType, certName, nextVersion, caCertSN, keyInfo.CertKeyLevel, keyInfo.CertKeyId, certData, cert.NotAfter.UnixMilli(), now.UnixMilli(), now.UnixMilli())
	if err != nil {
		return "", err
	}
	if n, err := r.RowsAffected(); err != nil {
		return "", err
	} else if n != 1 {
		return "", errors.New("number of rows affected is not 1")
	}

	success = true
	return certSN, nil
}

func (p *PostgresKeyStorage) SaveRootCA(certName string, cert *x509.Certificate, keyInfo secretapi.CertKeyInfo) (secretapi.CertSerialNumber, error) {
	return p.saveInternal(secretapi.CertLevelTypeRootCA, certName, "", cert, keyInfo)
}

func (p *PostgresKeyStorage) SaveIntermediateCA(certName string, caCertSerialNumber secretapi.CertSerialNumber, cert *x509.Certificate, keyInfo secretapi.CertKeyInfo) (secretapi.CertSerialNumber, error) {
	return p.saveInternal(secretapi.CertLevelTypeIntermediateCA, certName, caCertSerialNumber, cert, keyInfo)
}

func (p *PostgresKeyStorage) SaveCert(certName string, caCertSerialNumber secretapi.CertSerialNumber, cert *x509.Certificate, keyInfo secretapi.CertKeyInfo) (secretapi.CertSerialNumber, error) {
	return p.saveInternal(secretapi.CertLevelTypeCert, certName, caCertSerialNumber, cert, keyInfo)
}

func (p *PostgresKeyStorage) LoadCertByName(certName string, certLevelType secretapi.CertLevelType) (*x509.Certificate, secretapi.CertKeyInfo, error) {
	row := p.db.QueryRow(`select cert_content, cert_key_level, cert_key_name from secret_cert where cert_type = $1 and cert_name = $2 order by cert_version desc limit 1`, certLevelType, certName)
	if row == nil {
		return nil, secretapi.CertKeyInfo{}, errors.New("cert row is nil")
	}
	if row.Err() != nil {
		return nil, secretapi.CertKeyInfo{}, row.Err()
	}

	var certData []byte
	var keyInfo secretapi.CertKeyInfo
	if err := row.Scan(&certData, &keyInfo.CertKeyLevel, &keyInfo.CertKeyId); err != nil {
		return nil, secretapi.CertKeyInfo{}, err
	}

	cert, err := new(secretapi.PemTool).ParseCertificate(certData)
	if err != nil {
		return nil, secretapi.CertKeyInfo{}, err
	}
	return cert, keyInfo, nil
}

func (p *PostgresKeyStorage) LoadCertById(certSerialNumber secretapi.CertSerialNumber) (*x509.Certificate, secretapi.CertLevelType, secretapi.CertKeyInfo, error) {
	row := p.db.QueryRow(`select cert_type, cert_content, cert_key_level, cert_key_name from secret_cert where cert_id = $1 order by cert_version desc limit 1`, certSerialNumber)
	if row == nil {
		return nil, 0, secretapi.CertKeyInfo{}, errors.New("cert row is nil")
	}
	if row.Err() != nil {
		return nil, 0, secretapi.CertKeyInfo{}, row.Err()
	}

	var certData []byte
	var keyInfo secretapi.CertKeyInfo
	var certLevelType secretapi.CertLevelType
	if err := row.Scan(&certLevelType, &certData, &keyInfo.CertKeyLevel, &keyInfo.CertKeyId); err != nil {
		return nil, 0, secretapi.CertKeyInfo{}, err
	}

	cert, err := new(secretapi.PemTool).ParseCertificate(certData)
	if err != nil {
		return nil, 0, secretapi.CertKeyInfo{}, err
	}
	return cert, certLevelType, keyInfo, nil
}

func (p *PostgresKeyStorage) LoadParentCertByCertId(currentCertSerialNumber secretapi.CertSerialNumber) (*x509.Certificate, secretapi.CertLevelType, secretapi.CertKeyInfo, error) {
	row := p.db.QueryRow(`select parent_ca_cert_id from secret_cert where cert_id = $1 order by cert_version desc limit 1`, currentCertSerialNumber)
	if row == nil {
		return nil, 0, secretapi.CertKeyInfo{}, errors.New("cert row is nil")
	}
	if row.Err() != nil {
		return nil, 0, secretapi.CertKeyInfo{}, row.Err()
	}

	var parentCertId string
	if err := row.Scan(&parentCertId); err != nil {
		return nil, 0, secretapi.CertKeyInfo{}, err
	}
	if parentCertId == "" {
		return nil, 0, secretapi.CertKeyInfo{}, nil
	}

	return p.LoadCertById(secretapi.CertSerialNumber(parentCertId))
}

func (p *PostgresKeyStorage) LoadCertChainByName(certName string, certLevelType secretapi.CertLevelType) ([]*x509.Certificate, secretapi.CertKeyInfo, error) {
	cert, info, err := p.LoadCertByName(certName, certLevelType)
	if err != nil {
		return nil, secretapi.CertKeyInfo{}, err
	}

	var certs = []*x509.Certificate{cert}
	const MaxLevels = 10
	var prevCert *x509.Certificate = cert
	for i := 0; i < MaxLevels; i++ {
		var sn secretapi.CertSerialNumber
		sn.FromBigInt(prevCert.SerialNumber)
		cert, _, _, err := p.LoadParentCertByCertId(sn)
		if err != nil {
			return nil, secretapi.CertKeyInfo{}, err
		}
		if cert == nil {
			// reaching root
			return certs, info, nil
		}
		prevCert = cert
		certs = append(certs, cert)
	}
	return nil, secretapi.CertKeyInfo{}, errors.New("reaching max cert level")
}

func (p *PostgresKeyStorage) NextCertSerialNumber() (secretapi.CertSerialNumber, error) {
	row := p.db.QueryRow("select nextval('cert_id_seq')")
	if row == nil {
		return "", errors.New("no rows found")
	}
	if row.Err() != nil {
		return "", row.Err()
	}
	var id int64
	if err := row.Scan(&id); err != nil {
		return "", err
	}
	var num secretapi.CertSerialNumber
	num.FromInt64(id)
	return num, nil
}
