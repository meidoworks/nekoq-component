package secretapi

import (
	"crypto/x509"
	"encoding/hex"
	"math/big"
)

type CertKeyLevel int

const (
	CertKeyLevelUnseal CertKeyLevel = iota + 1
)

const (
	CertKeyLevelLevel1Rsa CertKeyLevel = iota + 1001
	CertKeyLevelLevel1Ecdsa
)

const (
	CertKeyLevelLevel2Rsa CertKeyLevel = iota + 2001
	CertKeyLevelLevel2Ecdsa
	CertKeyLevelLevel2Custom CertKeyLevel = 2999
)

const (
	CertKeyLevelExternal CertKeyLevel = iota + 9001
)

type CertLevelType int

const (
	CertLevelTypeRootCA CertLevelType = iota + 1
	CertLevelTypeIntermediateCA
	CertLevelTypeCert
)

type CertSerialNumber string

func (s CertSerialNumber) ToBigInt() (*big.Int, error) {
	data, err := hex.DecodeString(string(s))
	if err != nil {
		return nil, err
	}
	return new(big.Int).SetBytes(data), nil
}

func (s *CertSerialNumber) FromInt64(n int64) {
	data := big.NewInt(n).Bytes()
	*s = CertSerialNumber(hex.EncodeToString(data))
}

func (s *CertSerialNumber) FromBytes(data []byte) {
	*s = CertSerialNumber(hex.EncodeToString(data))
}

func (s *CertSerialNumber) FromBigInt(i *big.Int) {
	s.FromBytes(i.Bytes())
}

type CertKeyInfo struct {
	CertKeyLevel CertKeyLevel
	CertKeyId    string
}

type CertStorage interface {
	SaveRootCA(certName string, cert *x509.Certificate, keyInfo CertKeyInfo) (CertSerialNumber, error)
	SaveIntermediateCA(certName string, caCertSerialNumber CertSerialNumber, cert *x509.Certificate, keyInfo CertKeyInfo) (CertSerialNumber, error)
	SaveCert(certName string, caCertSerialNumber CertSerialNumber, cert *x509.Certificate, keyInfo CertKeyInfo) (CertSerialNumber, error)

	LoadCertByName(certName string, certLevelType CertLevelType) (*x509.Certificate, CertKeyInfo, error)
	LoadCertById(certSerialNumber CertSerialNumber) (*x509.Certificate, CertLevelType, CertKeyInfo, error)
	LoadParentCertByCertId(currentCertSerialNumber CertSerialNumber) (*x509.Certificate, CertLevelType, CertKeyInfo, error)

	NextCertSerialNumber() (CertSerialNumber, error)
}
