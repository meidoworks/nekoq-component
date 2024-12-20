package secretimpl_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/meidoworks/nekoq-component/configure/secretapi"
	"github.com/meidoworks/nekoq-component/configure/secretimpl"
)

func TestCertStoreNextSerialNumber(t *testing.T) {
	ks := InitTestKeyStorage(t).(*secretimpl.PostgresKeyStorage)
	serialNumber, err := ks.NextCertSerialNumber()
	if err != nil {
		t.Fatal(err)
	}
	if serialNumber == "" {
		t.Fatal("serial number should not be empty")
	}
	t.Log(serialNumber)
}

func TestCertStore1(t *testing.T) {
	ks := InitTestKeyStorage(t).(*secretimpl.PostgresKeyStorage)

	tool := secretapi.NewLevel2CipherTool(ks, secretapi.DefaultKeyGen, "test_case")
	if err := tool.NewRsaKey("test_cert_rsa_key", secretapi.KeyRSA4096); err != nil {
		t.Fatal(err)
	}

	keyId, keyType, key, err := ks.FetchL2DataKey("test_cert_rsa_key")
	if err != nil {
		t.Fatal(err)
	}
	if keyType != secretapi.KeyRSA4096 {
		t.Fatal("wrong key type")
	}
	pri, err := new(secretapi.PemTool).ParseRsaPrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("private key id:", keyId)
	newSnNumber, err := ks.NextCertSerialNumber()
	if err != nil {
		t.Fatal(err)
	}
	bigIntSnNumber, err := newSnNumber.ToBigInt()
	if err != nil {
		t.Fatal(err)
	}

	cert, err := new(secretapi.CertTool).CreateRootCACertificate((&secretapi.CACertReq{
		SerialNumber:  bigIntSnNumber,
		CommonName:    "Secret Test Root CA",
		Organization:  "MeidoWorks",
		Country:       "XX",
		Province:      "The State",
		Locality:      "The City",
		StreetAddress: "Address Street",
		PostalCode:    "100000",
		StartTime:     time.Now(),
		Version:       0,
	}).Duration(8760*time.Hour), new(secretapi.CertKeyPair).FromPrivateKey(pri))
	if err != nil {
		t.Fatal(err)
	}

	sn, err := ks.SaveRootCA("test_cert_root_ca", cert, secretapi.CertKeyInfo{
		CertKeyLevel: secretapi.CertKeyLevelLevel2Custom,
		CertKeyId:    fmt.Sprint(keyId),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("cert serial number:", sn)

	cert, keyInfo, err := ks.LoadCertByName("test_cert_root_ca", secretapi.CertLevelTypeRootCA)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("keyinfo key id:", keyInfo.CertKeyId)
	t.Log("keyinfo key level:", keyInfo.CertKeyLevel)
	t.Log("cert:", cert.Subject.CommonName)

	cert, levelType, keyInfo, err := ks.LoadCertById(sn)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("keyinfo key id:", keyInfo.CertKeyId)
	t.Log("keyinfo key level:", keyInfo.CertKeyLevel)
	t.Log("cert level type:", levelType)
	t.Log("cert:", cert.Subject.CommonName)

	var snNumber secretapi.CertSerialNumber
	snNumber.FromBigInt(cert.SerialNumber)
	cert, levelType, keyInfo, err = ks.LoadCertById(snNumber)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("keyinfo key id:", keyInfo.CertKeyId)
	t.Log("keyinfo key level:", keyInfo.CertKeyLevel)
	t.Log("cert level type:", levelType)
	t.Log("cert:", cert.Subject.CommonName)
}
