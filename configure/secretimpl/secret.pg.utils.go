package secretimpl

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/meidoworks/nekoq-component/configure/secretapi"
)

func checkAvailableKeyName(name string) error {
	if name == secretapi.TokenName {
		return errors.New("cannot use token name")
	}
	return nil
}

func checkRotateLevel2Key(keyType, expectedKeyType secretapi.KeyType) error {
	if keyType == 0 {
		return errors.New("key type is 0")
	}
	if keyType != expectedKeyType {
		return errors.New("key type mismatch")
	}
	return nil
}

func keySetEncrypt(keySet *secretapi.KeySet, up secretapi.UnsealProvider) ([]byte, string, error) {
	rawData, err := json.Marshal(keySet)
	if err != nil {
		return nil, "", err
	}
	return up.Encrypt(context.Background(), rawData)
}

func keySetDecrypt(data []byte, up secretapi.UnsealProvider) (*secretapi.KeySet, error) {
	rawData, err := up.Decrypt(context.Background(), data)
	if err != nil {
		return nil, err
	}
	keySet := &secretapi.KeySet{}
	err = json.Unmarshal(rawData, keySet)
	if err != nil {
		return nil, err
	}
	return keySet, nil
}

func keySetEncryptByL1(keySet *secretapi.KeySet, id int64, lv1Ks *secretapi.KeySet) ([]byte, error) {
	data, err := json.Marshal(keySet)
	if err != nil {
		return nil, err
	}
	ciphertext, nonce, err := lv1Ks.AesGCMEnc(data)
	if err != nil {
		return nil, err
	}
	encrypted := fmt.Sprintf("$%d$%s$%s", id, base64.StdEncoding.EncodeToString(ciphertext), base64.StdEncoding.EncodeToString(nonce))
	return []byte(encrypted), nil
}

func keySetDecryptByL1(data []byte, lv1Ks *secretapi.KeySet) (int64, *secretapi.KeySet, error) {
	var keyId int64
	var rest string
	if _, err := fmt.Sscanf(string(data), "$%d$%s", &keyId, &rest); err != nil {
		return 0, nil, err
	}
	splits := strings.Split(rest, "$")
	if len(splits) != 2 {
		return 0, nil, errors.New("invalid cipher data")
	}
	ciphertext, err := base64.StdEncoding.DecodeString(splits[0])
	if err != nil {
		return 0, nil, err
	}
	nonce, err := base64.StdEncoding.DecodeString(splits[1])
	if err != nil {
		return 0, nil, err
	}
	plaintext, err := lv1Ks.AesGCMDec(ciphertext, nonce)
	if err != nil {
		return 0, nil, err
	}
	keySet := &secretapi.KeySet{}
	err = json.Unmarshal(plaintext, keySet)
	if err != nil {
		return 0, nil, err
	}
	return keyId, keySet, nil
}
