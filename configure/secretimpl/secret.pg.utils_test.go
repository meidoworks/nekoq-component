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
