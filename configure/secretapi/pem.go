package secretapi

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

type PemTool struct {
}

func NewPemTool() *PemTool {
	return &PemTool{}
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
