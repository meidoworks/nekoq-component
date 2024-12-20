package secretapi

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

// PemTool supports encoding and decoding pem cert and private keys
// Pending items: password private key, bundled cert file
type PemTool struct {
}

func NewPemTool() *PemTool {
	return &PemTool{}
}

func (p *PemTool) EncodeKeySet(ks *KeySet) ([]byte, error) {
	data, err := ks.SaveAsBytes()
	if err != nil {
		return nil, err
	}
	ksBlock := &pem.Block{
		Type:  "KeySet",
		Bytes: data,
	}
	pemData := pem.EncodeToMemory(ksBlock)
	return pemData, nil
}

func (p *PemTool) ParseKeySet(data []byte) (*KeySet, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("pem decode failed")
	}
	ks := &KeySet{}
	if err := ks.LoadFromBytes(block.Bytes); err != nil {
		return nil, err
	}
	return ks, nil
}

func (p *PemTool) EncodeCertificateRevocationList(csr []byte) ([]byte, error) {
	certBlock := &pem.Block{
		Type:  "X509 CRL",
		Bytes: csr,
	}
	pemData := pem.EncodeToMemory(certBlock)
	return pemData, nil
}

func (p *PemTool) ParseCertificateRevocationList(pemData []byte) (*x509.RevocationList, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("pem decode failed")
	}
	req, err := x509.ParseRevocationList(block.Bytes)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func (p *PemTool) EncodeCertificateRequest(csr []byte) ([]byte, error) {
	certBlock := &pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: csr,
	}
	pemData := pem.EncodeToMemory(certBlock)
	return pemData, nil
}

func (p *PemTool) ParseCertificateRequest(pemData []byte) (*x509.CertificateRequest, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("pem decode failed")
	}
	req, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func (p *PemTool) EncodeCertificate(cert *x509.Certificate) ([]byte, error) {
	certBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	}
	pemData := pem.EncodeToMemory(certBlock)
	return pemData, nil
}

func (p *PemTool) ParseCertificate(data []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}
	return cert, nil
}

func (p *PemTool) EncodeRsaPrivateKey(pri *rsa.PrivateKey) ([]byte, error) {
	data := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(pri),
	})
	return data, nil
}

func (p *PemTool) ParseRsaPrivateKey(data []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("rsa private key error")
	}
	pri, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return pri, nil
}

func (p *PemTool) EncodeECDSAPrivateKey(pri *ecdsa.PrivateKey) ([]byte, error) {
	keydata, err := x509.MarshalECPrivateKey(pri)
	if err != nil {
		return nil, err
	}
	data := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: keydata,
	})
	return data, nil
}

func (p *PemTool) ParseECDSAPrivateKey(data []byte) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("ecdsa private key error")
	}
	pri, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return pri, nil
}
