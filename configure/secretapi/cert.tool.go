package secretapi

import (
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"net/url"
	"time"
)

// CertKeyPair is used for issuing new certificates
// If the key pair is from an external provider, crypto.PrivateKey and crypto.PublicKey interfaces can be implemented.
type CertKeyPair struct {
	PrivateKey crypto.Signer
	PublicKey  crypto.PublicKey
}

func (k *CertKeyPair) FromPrivateKey(pri crypto.Signer) *CertKeyPair {
	k.PrivateKey = pri
	k.PublicKey = pri.Public()
	return k
}

type CACertReq struct {
	SerialNumber *big.Int

	CommonName    string
	Organization  string
	Country       string
	Province      string
	Locality      string
	StreetAddress string
	PostalCode    string

	StartTime  time.Time
	ExpireTime time.Time
	Version    int
}

func (k *CACertReq) Duration(duration time.Duration) *CACertReq {
	k.ExpireTime = k.StartTime.Add(duration)
	return k
}

type CertReq struct {
	CommonName    string
	Organization  string
	Country       string
	Province      string
	Locality      string
	StreetAddress string
	PostalCode    string

	DNSNames       []string
	EmailAddresses []string
	IPAddresses    []net.IP
	URLs           []*url.URL
}

type CertMeta struct {
	SerialNumber *big.Int
	StartTime    time.Time
	ExpireTime   time.Time
	Version      int
	SignerCert   *x509.Certificate
	Signer       *CertKeyPair
}

func (m *CertMeta) Duration(duration time.Duration) *CertMeta {
	m.ExpireTime = m.StartTime.Add(duration)
	return m
}

type CertTool struct {
	EndUserCertKeyUsage    x509.KeyUsage
	EndUserCertExtKeyUsage []x509.ExtKeyUsage
}

func (c *CertTool) SetupDefaultServerCertKeyUsage() *CertTool {
	c.EndUserCertKeyUsage = x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageDataEncipherment | x509.KeyUsageKeyAgreement
	c.EndUserCertExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
	return c
}

func (c *CertTool) createCACert(req *CACertReq, caCert *x509.Certificate, certKeyPair, signerKeyPair *CertKeyPair, maxPath int) (*x509.Certificate, error) {
	// max path length settings
	maxPathZero := maxPath == 0
	if maxPath == -1 {
		maxPath = 0
	}

	cert := &x509.Certificate{
		SerialNumber: req.SerialNumber,
		Subject: pkix.Name{
			Country:      []string{req.Country},
			Organization: []string{req.Organization},
			//TODO OrganizationUnit
			Locality:      []string{req.Locality},
			Province:      []string{req.Province},
			StreetAddress: []string{req.StreetAddress},
			PostalCode:    []string{req.PostalCode},
			CommonName:    req.CommonName,
		},
		NotBefore:             req.StartTime,
		NotAfter:              req.ExpireTime,
		IsCA:                  true,
		Version:               req.Version,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign | x509.KeyUsageDigitalSignature,
		// no ExtKeyUsage specified
		ExtKeyUsage: nil,
		// just allow unlimited path length
		MaxPathLen:     maxPath,
		MaxPathLenZero: maxPathZero,
	}
	// if no ca cert specified, use self-signed cert
	if caCert == nil {
		caCert = cert
	}

	certDerData, err := x509.CreateCertificate(rand.Reader, cert, caCert, certKeyPair.PublicKey, signerKeyPair.PrivateKey)
	if err != nil {
		return nil, err
	}
	outputCert, err := x509.ParseCertificate(certDerData)
	if err != nil {
		return nil, err
	}
	return outputCert, nil
}

func (c *CertTool) CreateRootCACertificate(req *CACertReq, keyPair *CertKeyPair) (*x509.Certificate, error) {
	return c.createCACert(req, nil, keyPair, keyPair, -1)
}

func (c *CertTool) CreateIntermediateCACertificate(req *CACertReq, ca *x509.Certificate, caKeyPair, curKeyPair *CertKeyPair) (*x509.Certificate, error) {
	return c.createCACert(req, ca, curKeyPair, caKeyPair, 0)
}

func (c *CertTool) CreateCertificateRequest(req *CertReq, curKeyPair *CertKeyPair) (*x509.CertificateRequest, error) {
	result := &x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:    req.CommonName,
			Organization:  []string{req.Organization},
			Country:       []string{req.Country},
			Province:      []string{req.Province},
			Locality:      []string{req.Locality},
			StreetAddress: []string{req.StreetAddress},
			PostalCode:    []string{req.PostalCode},
		},
		DNSNames:       req.DNSNames,
		EmailAddresses: req.EmailAddresses,
		IPAddresses:    req.IPAddresses,
		URIs:           req.URLs,

		Version: 0,
	}

	csr, err := x509.CreateCertificateRequest(rand.Reader, result, curKeyPair.PrivateKey)
	if err != nil {
		return nil, err
	}
	newResult, err := x509.ParseCertificateRequest(csr)
	if err != nil {
		return nil, err
	}

	return newResult, nil
}

func (c *CertTool) CreateCertificate(req *x509.CertificateRequest, meta *CertMeta) (*x509.Certificate, error) {
	cert := &x509.Certificate{
		SerialNumber:          meta.SerialNumber,
		Subject:               req.Subject,
		NotBefore:             meta.StartTime,
		NotAfter:              meta.ExpireTime,
		Version:               req.Version,
		IsCA:                  false,
		BasicConstraintsValid: true,

		DNSNames:       req.DNSNames,
		EmailAddresses: req.EmailAddresses,
		IPAddresses:    req.IPAddresses,
		URIs:           req.URIs,

		KeyUsage:    c.EndUserCertKeyUsage,
		ExtKeyUsage: c.EndUserCertExtKeyUsage,
	}

	certDerData, err := x509.CreateCertificate(rand.Reader, cert, meta.SignerCert, req.PublicKey, meta.Signer.PrivateKey)
	if err != nil {
		return nil, err
	}
	outputCert, err := x509.ParseCertificate(certDerData)
	if err != nil {
		return nil, err
	}
	return outputCert, nil
}
