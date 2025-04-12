package secretserver

import (
	"crypto"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	"github.com/meidoworks/nekoq-component/configure/permissions"
	"github.com/meidoworks/nekoq-component/configure/secretaddon"
	"github.com/meidoworks/nekoq-component/configure/secretapi"
	"github.com/meidoworks/nekoq-component/http/chi2"
)

// CertManageCreateCert generates certificates for the request
type CertManageCreateCert struct {
	chi2.Controller
	Method string `method:"POST"`
	URL    string `url:"/api/v1/secret/cert/new"`

	cert       secretapi.CertStorage
	keyStorage secretapi.KeyStorage
}

type createCertReq struct {
	CAName      string `json:"ca_name"` // ca cert should have level2 key
	TTL         int    `json:"ttl"`
	CertName    string `json:"cert_name"`
	CertReqData string `json:"cert_req_data"`
	CertUsage   string `json:"cert_usage"` // available: server/client/both
	//FIXME support managed private key
}

func (c *CertManageCreateCert) HandleHttp(w http.ResponseWriter, r *http.Request) chi2.Render {
	obj, err := c.ParseBody(r)
	if err != nil {
		return chi2.NewErrRender(err)
	}
	req := obj.(*createCertReq)
	certName := strings.TrimSpace(req.CertName)
	certReqData := req.CertReqData
	caName := strings.TrimSpace(req.CAName)
	if certName == "" || certReqData == "" || caName == "" {
		return chi2.NewStatusRender(http.StatusBadRequest)
	}
	if req.CertUsage != "server" && req.CertUsage != "client" && req.CertUsage != "both" {
		return chi2.NewStatusRender(http.StatusBadRequest)
	}
	pemtool := new(secretapi.PemTool)

	// prepare cert request
	certReq, err := new(secretapi.PemTool).ParseCertificateRequest([]byte(certReqData))
	if err != nil {
		return chi2.NewErrRender(err)
	}

	// prepare ca cert and ca key
	caCert, caKeyInfo, err := c.cert.LoadCertByName(caName, secretapi.CertLevelTypeIntermediateCA)
	if err != nil {
		return chi2.NewErrRender(err)
	}
	certKeyIdNum, err := strconv.ParseInt(caKeyInfo.CertKeyId, 10, 64)
	if err != nil {
		return chi2.NewErrRender(err)
	}
	var signer crypto.Signer
	switch caKeyInfo.CertKeyLevel {
	case secretapi.CertKeyLevelLevel2Custom:
		caKeyType, caKey, err := c.keyStorage.LoadL2DataKeyById(certKeyIdNum)
		if err != nil {
			return chi2.NewErrRender(err)
		}
		switch caKeyType {
		case secretapi.KeyRSA1024, secretapi.KeyRSA2048, secretapi.KeyRSA3072, secretapi.KeyRSA4096:
			pk, err := pemtool.ParseRsaPrivateKey(caKey)
			if err != nil {
				return chi2.NewErrRender(err)
			}
			signer = pk
		case secretapi.KeyECDSA224, secretapi.KeyECDSA256, secretapi.KeyECDSA384, secretapi.KeyECDSA521:
			pk, err := pemtool.ParseECDSAPrivateKey(caKey)
			if err != nil {
				return chi2.NewErrRender(err)
			}
			signer = pk
		default:
			return chi2.NewErrRender(errors.New("unsupported key type"))
		}
	case secretapi.CertKeyLevelLevel2Rsa:
		keySet, err := c.keyStorage.LoadLevel2KeySetById(certKeyIdNum)
		if err != nil {
			return chi2.NewErrRender(err)
		}
		pk, err := keySet.Rsa()
		if err != nil {
			return chi2.NewErrRender(err)
		}
		signer = pk
	case secretapi.CertKeyLevelLevel2Ecdsa:
		keySet, err := c.keyStorage.LoadLevel2KeySetById(certKeyIdNum)
		if err != nil {
			return chi2.NewErrRender(err)
		}
		pk, _, err := keySet.Ecdsa()
		if err != nil {
			return chi2.NewErrRender(err)
		}
		signer = pk
	default:
		return chi2.NewErrRender(errors.New("unsupported cert key level"))
	}

	var caCertSnNum secretapi.CertSerialNumber
	caCertSnNum.FromBigInt(caCert.SerialNumber)
	caCerts, err := c.recursiveRetrieveCerts(caCert, caCertSnNum)
	if err != nil {
		return chi2.NewErrRender(err)
	}
	var caCertData []string
	for _, cert := range caCerts {
		data, err := pemtool.EncodeCertificate(cert)
		if err != nil {
			return chi2.NewErrRender(err)
		}
		caCertData = append(caCertData, string(data))
	}

	// generate next serial number
	certSnNum, err := c.cert.NextCertSerialNumber()
	if err != nil {
		return chi2.NewErrRender(err)
	}
	certSnBig, err := certSnNum.ToBigInt()
	if err != nil {
		return chi2.NewErrRender(err)
	}

	// create cert
	tool := new(secretapi.CertTool)
	switch req.CertUsage {
	case "server":
		tool.SetupDefaultServerCertKeyUsage()
	case "client":
		tool.SetupDefaultClientCertKeyUsage()
	case "both":
		tool.SetupDefaultBothCertKeyUsage()
	default:
		return chi2.NewErrRender(errors.New("unsupported cert usage"))
	}
	newCert, err := tool.CreateCertificate(certReq, (&secretapi.CertMeta{
		SerialNumber: certSnBig,
		StartTime:    time.Now(),
		SignerCert:   caCert,
		Signer: &secretapi.CertKeyPair{
			PrivateKey: signer,
			PublicKey:  signer.Public(),
		},
	}).Duration(time.Duration(req.TTL)*time.Second))
	newCertData, err := pemtool.EncodeCertificate(newCert)
	if err != nil {
		return chi2.NewErrRender(err)
	}

	// save cert
	//FIXME support managed private key
	if _, err := c.cert.SaveCert(certName, caCertSnNum, newCert, secretapi.CertKeyInfo{
		CertKeyLevel: secretapi.CertKeyLevelExternal,
	}); err != nil {
		return chi2.NewErrRender(err)
	}

	// respond cert + ca chain
	return chi2.NewJsonOkRender(map[string]interface{}{
		"cert":    string(newCertData),
		"ca_list": caCertData,
	})
}

func (c *CertManageCreateCert) recursiveRetrieveCerts(cert *x509.Certificate, sn secretapi.CertSerialNumber) ([]*x509.Certificate, error) {
	pcert, _, _, err := c.cert.LoadParentCertByCertId(sn)
	if err != nil {
		return nil, err
	}
	if pcert == nil {
		return []*x509.Certificate{cert}, nil
	}

	var nextSnNum secretapi.CertSerialNumber
	nextSnNum.FromBigInt(pcert.SerialNumber)
	results, err := c.recursiveRetrieveCerts(pcert, nextSnNum)
	if err != nil {
		return nil, err
	}
	return append(results, cert), nil
}

func NewCertManageCreateCert(verifier secretapi.JwtVerifier, keyStorage secretapi.KeyStorage, certStorage secretapi.CertStorage) *CertManageCreateCert {
	return &CertManageCreateCert{
		Controller: chi2.Controller{
			Middlewares: chi2.Middlewares{
				middleware.Logger,
				middleware.RealIP,
				middleware.Recoverer,
			},
			RequestValidators: chi2.RequestValidators{
				chi2.AllowContentTypeFor("application/json"),
				ValidateJwtToken(verifier, secretaddon.PermissionResourceList{}.
					Add(permissions.SecretCertAdmin)),
			},
			BodyParser: func(r io.Reader) (any, error) {
				req := new(createCertReq)
				if err := render.DecodeJSON(r, req); err != nil {
					return nil, err
				}
				return req, nil
			},
		},
		keyStorage: keyStorage,
		cert:       certStorage,
	}
}

// CertManageCreateCertReq generates certificate requests using builtin private keys
type CertManageCreateCertReq struct {
	chi2.Controller
	Method string `method:"POST"`
	URL    string `url:"/api/v1/secret/cert/newreq"`

	keyStorage secretapi.KeyStorage
}

type createCertReqReq struct {
	KeyName string `json:"key_name"` // the key should be L2 key

	CommonName     string   `json:"cn"`
	Organization   string   `json:"org"`
	Country        string   `json:"country"`
	Province       string   `json:"province"`
	Locality       string   `json:"locality"`
	StreetAddress  string   `json:"street_address"`
	PostalCode     string   `json:"postal_code"`
	DNSNames       []string `json:"dns_names"`
	EmailAddresses []string `json:"email_addresses"`
}

func (c *CertManageCreateCertReq) HandleHttp(w http.ResponseWriter, r *http.Request) chi2.Render {
	obj, err := c.ParseBody(r)
	if err != nil {
		return chi2.NewErrRender(err)
	}
	req := obj.(*createCertReqReq)
	if req.KeyName == "" {
		return chi2.NewStatusRender(http.StatusBadRequest)
	}

	keyId, keyType, key, err := c.keyStorage.FetchL2DataKey(req.KeyName)
	if err != nil {
		return chi2.NewErrRender(err)
	}
	var signer crypto.Signer
	switch keyType {
	case secretapi.KeyRSA1024, secretapi.KeyRSA2048, secretapi.KeyRSA4096, secretapi.KeyRSA3072:
		pk, err := new(secretapi.PemTool).ParseRsaPrivateKey(key)
		if err != nil {
			return chi2.NewErrRender(err)
		}
		signer = pk
	case secretapi.KeyECDSA224, secretapi.KeyECDSA256, secretapi.KeyECDSA384, secretapi.KeyECDSA521:
		pk, err := new(secretapi.PemTool).ParseECDSAPrivateKey(key)
		if err != nil {
			return chi2.NewErrRender(err)
		}
		signer = pk
	default:
		return chi2.NewErrRender(errors.New("unsupported key type"))
	}

	certTool := new(secretapi.CertTool)
	certReq, err := certTool.CreateCertificateRequest(&secretapi.CertReq{
		CommonName:     req.CommonName,
		Organization:   req.Organization,
		Country:        req.Country,
		Province:       req.Province,
		Locality:       req.Locality,
		StreetAddress:  req.StreetAddress,
		PostalCode:     req.PostalCode,
		DNSNames:       req.DNSNames,
		EmailAddresses: req.EmailAddresses,
		IPAddresses:    nil, //FIXME support ip addresses field
		URLs:           nil, //FIXME support URLs field
	}, &secretapi.CertKeyPair{
		PrivateKey: signer,
		PublicKey:  signer.Public(),
	})
	if err != nil {
		return chi2.NewErrRender(err)
	}
	result, err := new(secretapi.PemTool).EncodeCertificateRequest(certReq.Raw)
	if err != nil {
		return chi2.NewErrRender(err)
	}

	return chi2.NewJsonOkRender(map[string]interface{}{
		"key_id": fmt.Sprint(keyId),
		"req":    string(result),
	})
}

func NewCertManageCreateCertReq(verifier secretapi.JwtVerifier, keyStorage secretapi.KeyStorage) *CertManageCreateCertReq {
	return &CertManageCreateCertReq{
		Controller: chi2.Controller{
			Middlewares: chi2.Middlewares{
				middleware.Logger,
				middleware.RealIP,
				middleware.Recoverer,
			},
			RequestValidators: chi2.RequestValidators{
				chi2.AllowContentTypeFor("application/json"),
				ValidateJwtToken(verifier, secretaddon.PermissionResourceList{}.
					Add(permissions.SecretCertAdmin)),
			},
			BodyParser: func(r io.Reader) (any, error) {
				req := new(createCertReqReq)
				if err := render.DecodeJSON(r, req); err != nil {
					return nil, err
				}
				return req, nil
			},
		},
		keyStorage: keyStorage,
	}
}

// CertManageGetCert gets cert, private key and ca if all maintained by secret
type CertManageGetCert struct {
	chi2.Controller
	Method string `method:"GET"`
	URL    string `url:"/api/v1/secret/cert/name/{cert_name}"`

	keyStorage  secretapi.KeyStorage
	certStorage secretapi.CertStorage
}

func (c *CertManageGetCert) HandleHttp(w http.ResponseWriter, r *http.Request) chi2.Render {
	certName := strings.TrimSpace(chi.URLParam(r, "cert_name"))
	if certName == "" {
		return chi2.NewStatusRender(http.StatusBadRequest)
	}

	certs, info, err := c.certStorage.LoadCertChainByName(certName, secretapi.CertLevelTypeCert)
	if err != nil {
		return chi2.NewErrRender(err)
	}

	certKeyIdNum, err := strconv.ParseInt(info.CertKeyId, 10, 64)
	if err != nil {
		return chi2.NewErrRender(err)
	}
	var certKey []byte
	switch info.CertKeyLevel {
	case secretapi.CertKeyLevelLevel2Custom:
		_, key, err := c.keyStorage.LoadL2DataKeyById(certKeyIdNum)
		if err != nil {
			return chi2.NewErrRender(err)
		}
		certKey = key
	case secretapi.CertKeyLevelLevel2Rsa:
		keySet, err := c.keyStorage.LoadLevel2KeySetById(certKeyIdNum)
		if err != nil {
			return chi2.NewErrRender(err)
		}
		certKey = keySet.RSA4096
	case secretapi.CertKeyLevelLevel2Ecdsa:
		keySet, err := c.keyStorage.LoadLevel2KeySetById(certKeyIdNum)
		if err != nil {
			return chi2.NewErrRender(err)
		}
		certKey = keySet.ECDSA_P521
	}

	pemtool := new(secretapi.PemTool)
	certStr, err := pemtool.EncodeCertificate(certs[0])
	if err != nil {
		return chi2.NewErrRender(err)
	}
	var caCerts []string
	for _, v := range certs[1:] {
		str, err := pemtool.EncodeCertificate(v)
		if err != nil {
			return chi2.NewErrRender(err)
		}
		caCerts = append(caCerts, string(str))
	}
	if len(certs) > 0 {
		return chi2.NewJsonOkRender(map[string]interface{}{
			"cert":    string(certStr),
			"ca_list": caCerts,
			"key":     string(certKey),
		})
	} else {
		return chi2.NewJsonOkRender(map[string]interface{}{
			"cert":    string(certStr),
			"ca_list": caCerts,
		})
	}
}

func NewCertManageGetCert(verifier secretapi.JwtVerifier, keyStorage secretapi.KeyStorage, certStorage secretapi.CertStorage) *CertManageGetCert {
	return &CertManageGetCert{
		Controller: chi2.Controller{
			Middlewares: chi2.Middlewares{
				middleware.Logger,
				middleware.RealIP,
				middleware.Recoverer,
			},
			RequestValidators: chi2.RequestValidators{
				chi2.AllowContentTypeFor("application/json"),
				ValidateJwtToken(verifier, secretaddon.PermissionResourceList{}.
					Add(permissions.SecretCertAdmin)),
			},
			BodyParser: func(r io.Reader) (any, error) {
				req := new(createCertReqReq)
				if err := render.DecodeJSON(r, req); err != nil {
					return nil, err
				}
				return req, nil
			},
		},

		keyStorage:  keyStorage,
		certStorage: certStorage,
	}
}
