package secretimpl_test

import (
	"slices"
	"testing"

	"github.com/meidoworks/nekoq-component/configure/secretapi"
	"github.com/meidoworks/nekoq-component/utility/random"
)

func TestAes128(t *testing.T) {
	keyStorage := InitTestKeyStorage(t)

	tool := secretapi.NewLevel2CipherTool(keyStorage, secretapi.DefaultKeyGen, "test_case")
	if err := tool.NewAes128Key("test_case_aes128"); err != nil {
		t.Fatal(err)
	}
}

func TestCipherToolAes128(t *testing.T) {
	keyStorage := InitTestKeyStorage(t)

	tool := secretapi.NewLevel2CipherTool(keyStorage, secretapi.DefaultKeyGen, "test_case")
	keyName := random.String(10)

	if err := tool.NewAes128Key(keyName); err != nil {
		t.Fatal(err)
	}

	raw := []byte{1, 2, 3, 4, 5, 6, 7, 8, 7, 6, 5, 4, 3, 2, 1}
	keyId, ciphertext, nonce, err := tool.Aes128Encrypt(keyName, raw, nil)
	if err != nil {
		t.Fatal(err)
	}

	plaintext, err := tool.Aes128Decrypt(keyId, ciphertext, nonce, nil)
	if err != nil {
		t.Fatal(err)
	}

	if slices.Compare(plaintext, raw) != 0 {
		t.Fatal("plaintext is not equal to raw")
	}
}
