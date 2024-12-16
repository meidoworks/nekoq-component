package secretapi

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"testing"
	"time"
)

var (
	caPri       crypto.Signer
	caCert      *x509.Certificate
	interCaPri  crypto.Signer
	interCaCert *x509.Certificate
)

func initCaKeys() {
	pri, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		panic(err)
	}
	caPri = pri
	pri2, err := rsa.GenerateKey(rand.Reader, 3072)
	if err != nil {
		panic(err)
	}
	interCaPri = pri2

	cert, err := new(CertTool).CreateRootCACertificate((&CACertReq{
		SerialNumber:  big.NewInt(time.Now().Unix()),
		CommonName:    "Secret Test Root CA",
		Organization:  "MeidoWorks",
		Country:       "XX",
		Province:      "The State",
		Locality:      "The City",
		StreetAddress: "Address Street",
		PostalCode:    "100000",
		StartTime:     time.Now(),
		Version:       0,
	}).Duration(8760*time.Hour), new(CertKeyPair).FromPrivateKey(pri))
	if err != nil {
		panic(err)
	}
	caCert = cert

	intermediateCert, err := new(CertTool).CreateIntermediateCACertificate((&CACertReq{
		SerialNumber:  big.NewInt(time.Now().Unix()),
		CommonName:    "Secret Test Intermediate CA",
		Organization:  "MeidoWorks",
		Country:       "XX",
		Province:      "The State",
		Locality:      "The City",
		StreetAddress: "Address Street",
		PostalCode:    "100000",
		StartTime:     time.Now(),
		Version:       0,
	}).Duration(8760*time.Hour), caCert, new(CertKeyPair).FromPrivateKey(caPri), new(CertKeyPair).FromPrivateKey(interCaPri))
	if err != nil {
		panic(err)
	}
	interCaCert = intermediateCert
}

func TestCaGenExample(t *testing.T) {
	initCaKeys()

	cert, err := new(CertTool).CreateRootCACertificate((&CACertReq{
		SerialNumber:  big.NewInt(time.Now().Unix()),
		CommonName:    "Secret Test Root CA",
		Organization:  "MeidoWorks",
		Country:       "XX",
		Province:      "The State",
		Locality:      "The City",
		StreetAddress: "Address Street",
		PostalCode:    "100000",
		StartTime:     time.Now(),
		ExpireTime:    time.Date(2050, time.December, 31, 23, 59, 59, 999_999_999, time.UTC),
		Version:       0,
	}).Duration(8760*time.Hour), new(CertKeyPair).FromPrivateKey(caPri))
	if err != nil {
		t.Fatal(err)
	}

	certBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	}
	pemData := pem.EncodeToMemory(certBlock)
	t.Log(string(pemData))
}

func TestIntermediateCaGenExample(t *testing.T) {
	initCaKeys()

	intermediateCert, err := new(CertTool).CreateIntermediateCACertificate((&CACertReq{
		SerialNumber:  big.NewInt(time.Now().Unix()),
		CommonName:    "Secret Test Intermediate CA",
		Organization:  "MeidoWorks",
		Country:       "XX",
		Province:      "The State",
		Locality:      "The City",
		StreetAddress: "Address Street",
		PostalCode:    "100000",
		StartTime:     time.Now(),
		ExpireTime:    time.Date(2050, time.December, 31, 23, 59, 59, 999_999_999, time.UTC),
		Version:       0,
	}).Duration(8760*time.Hour), caCert, new(CertKeyPair).FromPrivateKey(caPri), new(CertKeyPair).FromPrivateKey(interCaPri))
	if err != nil {
		t.Fatal(err)
	}

	pemData, err := new(PemTool).EncodeCertificate(intermediateCert)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(pemData))

	caPemData, err := new(PemTool).EncodeCertificate(caCert)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(caPemData))
}

func TestEndUserCertGenExample(t *testing.T) {
	initCaKeys()

	key, err := ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	req, err := new(CertTool).SetupDefaultServerCertKeyUsage().CreateCertificateRequest(&CertReq{
		CommonName:    "moetang.net",
		Organization:  "MeidoWorks Test",
		Country:       "CN",
		Province:      "Test Province",
		Locality:      "Test Locality",
		StreetAddress: "XXX Street",
		PostalCode:    "200000",
		DNSNames:      []string{"moetang.net", "moetang.com", "moetang.info", "moetang.org"},
	}, new(CertKeyPair).FromPrivateKey(key))
	if err != nil {
		t.Fatal(err)
	}
	cert, err := new(CertTool).SetupDefaultServerCertKeyUsage().CreateCertificate(req, (&CertMeta{
		SerialNumber: big.NewInt(time.Now().Unix()),
		StartTime:    time.Now(),
		SignerCert:   interCaCert,
		Signer:       new(CertKeyPair).FromPrivateKey(interCaPri),
	}).Duration(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	pemData, err := new(PemTool).EncodeCertificate(cert)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(pemData))

	interCaPemData, err := new(PemTool).EncodeCertificate(interCaCert)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(interCaPemData))

	caPemData, err := new(PemTool).EncodeCertificate(caCert)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(caPemData))
}
