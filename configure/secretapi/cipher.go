package secretapi

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
)

type Level2CipherTool struct {
	storage KeyStorage
	keyGen  KeyGen
	l1key   string
}

func NewLevel2CipherTool(storage KeyStorage, keyGen KeyGen, l1key string) *Level2CipherTool {
	return &Level2CipherTool{
		storage: storage,
		keyGen:  keyGen,
		l1key:   l1key,
	}
}

func (l *Level2CipherTool) internalNewKey(genFn func() ([]byte, error), name string, keyType KeyType) error {
	key, err := genFn()
	if err != nil {
		return err
	}
	return l.storage.StoreL2DataKey(l.l1key, name, keyType, key)
}

func (l *Level2CipherTool) NewGeneral64BKey(name string) error {
	return l.internalNewKey(l.keyGen.General64B, name, KeyGeneral64B)
}

func (l *Level2CipherTool) NewGeneral128BKey(name string) error {
	return l.internalNewKey(l.keyGen.General128B, name, KeyGeneral128B)
}

func (l *Level2CipherTool) NewAes128Key(name string) error {
	return l.internalNewKey(l.keyGen.Aes128, name, KeyAES128)
}

func (l *Level2CipherTool) NewRsaKey(name string, keyType KeyType) error {
	return l.internalNewKey(func() ([]byte, error) {
		return l.keyGen.Rsa(keyType)
	}, name, keyType)
}

func (l *Level2CipherTool) Aes128Encrypt(name string, plaintext, additionalData []byte) (keyId int64, rCiphertext, rNonce []byte, rerr error) {
	keyId, kt, key, err := l.storage.FetchL2DataKey(name)
	if err != nil {
		return 0, nil, nil, err
	}
	if kt != KeyAES128 {
		return 0, nil, nil, errors.New("key type mismatch")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return 0, nil, nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return 0, nil, nil, err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return 0, nil, nil, err
	}
	encData := aead.Seal(nil, nonce, plaintext, additionalData)
	return keyId, encData, nonce, nil
}

func (l *Level2CipherTool) Aes128Decrypt(keyId int64, ciphertext, nonce, additionalData []byte) (rPlaintext []byte, rerr error) {
	kt, key, err := l.storage.LoadL2DataKeyById(keyId)
	if err != nil {
		return nil, err
	}
	if kt != KeyAES128 {
		return nil, errors.New("key type mismatch")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return aead.Open(nil, nonce, ciphertext, additionalData)
}
