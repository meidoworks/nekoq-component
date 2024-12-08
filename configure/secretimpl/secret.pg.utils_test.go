package secretimpl

import (
	"testing"

	"github.com/meidoworks/nekoq-component/configure/secretapi"
)

func TestKeySetCipherByL1(t *testing.T) {
	lv1ks, err := secretapi.DefaultKeyGen.GenerateVitalKeySet()
	if err != nil {
		t.Fatal(err)
	}

	ks, err := secretapi.DefaultKeyGen.GenerateVitalKeySet()
	if err != nil {
		t.Fatal(err)
	}

	ciphertext, err := keySetEncryptByL1(ks, 123, lv1ks)
	if err != nil {
		t.Fatal(err)
	}
	keyId, newKs, err := keySetDecryptByL1(ciphertext, lv1ks)
	if err != nil {
		t.Fatal(err)
	}

	if keyId != 123 {
		t.Fatal("key id mismatch")
	}
	if ks.CRC_AES256 != newKs.CRC_AES256 {
		t.Fatal("key crc mismatch")
	}
}

func TestDataCipherByL1(t *testing.T) {

	lv1ks, err := secretapi.DefaultKeyGen.GenerateVitalKeySet()
	if err != nil {
		t.Fatal(err)
	}

	ciphertext, err := dataEncryptByL1([]byte{1, 2, 3, 4, 5, 6, 7, 8}, 123, lv1ks)
	if err != nil {
		t.Fatal(err)
	}
	keyId, newKs, err := dataDecryptByL1(ciphertext, lv1ks)
	if err != nil {
		t.Fatal(err)
	}

	if keyId != 123 {
		t.Fatal("key id mismatch")
	}
	if len(newKs) != 8 {
		t.Fatal("keys length mismatch")
	}
	if newKs[5] != 5+1 {
		t.Fatal("data mismatch")
	}
}
