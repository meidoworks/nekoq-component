package secretapi

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"hash/crc32"
)

type KeySet struct {
	AES256          []byte `json:"aes256"`
	CRC_AES256      string `json:"crc_aes256"`
	RSA4096         []byte `json:"rsa_4096"`
	CRC_RSA4096     string `json:"crc_rsa_4096"`
	ECDSA_P521      []byte `json:"ecdsa_p521"`
	CRC_ECDSA_P521  string `json:"crc_ecdsa_p521"`
	ED25519_PRI     []byte `json:"ed25519_pri"`
	CRC_ED25519_PRI string `json:"crc_ed25519_pri"`
	ED25519_PUB     []byte `json:"ed25519_pub"`
	CRC_ED25519_PUB string `json:"crc_ed25519_pub"`
}

func (k *KeySet) Aes() (cipher.Block, error) {
	return aes.NewCipher(k.AES256)
}

func (k *KeySet) AesGCMEnc(data []byte) ([]byte, []byte, error) {
	block, err := k.Aes()
	if err != nil {
		return nil, nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, err
	}
	encData := aead.Seal(nil, nonce, data, nil)
	return encData, nonce, nil
}

func (k *KeySet) AesGCMDec(encrypted, nonce []byte) ([]byte, error) {
	block, err := k.Aes()
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return aead.Open(nil, nonce, encrypted, nil)
}

func (k *KeySet) Rsa() (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(k.RSA4096)
	if block == nil {
		return nil, errors.New("rsa private key error")
	}
	pri, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return pri, nil
}

func (k *KeySet) Ecdsa() (*ecdsa.PrivateKey, *ecdsa.PublicKey, error) {
	block, _ := pem.Decode(k.ECDSA_P521)
	if block == nil {
		return nil, nil, errors.New("ecdsa private key error")
	}
	pri, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, nil, err
	}
	return pri, &pri.PublicKey, nil
}

func (k *KeySet) Ed25519() (ed25519.PrivateKey, ed25519.PublicKey, error) {
	return k.ED25519_PRI, k.ED25519_PUB, nil
}

func (k *KeySet) VerifyCrc() bool {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, crc32.ChecksumIEEE(k.AES256))
	if k.CRC_AES256 != hex.EncodeToString(buf) {
		return false
	}
	binary.LittleEndian.PutUint32(buf, crc32.ChecksumIEEE(k.RSA4096))
	if k.CRC_RSA4096 != hex.EncodeToString(buf) {
		return false
	}
	binary.LittleEndian.PutUint32(buf, crc32.ChecksumIEEE(k.ECDSA_P521))
	if k.CRC_ECDSA_P521 != hex.EncodeToString(buf) {
		return false
	}
	binary.LittleEndian.PutUint32(buf, crc32.ChecksumIEEE(k.ED25519_PRI))
	if k.CRC_ED25519_PRI != hex.EncodeToString(buf) {
		return false
	}
	binary.LittleEndian.PutUint32(buf, crc32.ChecksumIEEE(k.ED25519_PUB))
	if k.CRC_ED25519_PUB != hex.EncodeToString(buf) {
		return false
	}
	return true
}

func (k *KeySet) CalculateCrc() {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, crc32.ChecksumIEEE(k.AES256))
	k.CRC_AES256 = hex.EncodeToString(buf)
	binary.LittleEndian.PutUint32(buf, crc32.ChecksumIEEE(k.RSA4096))
	k.CRC_RSA4096 = hex.EncodeToString(buf)
	binary.LittleEndian.PutUint32(buf, crc32.ChecksumIEEE(k.ECDSA_P521))
	k.CRC_ECDSA_P521 = hex.EncodeToString(buf)
	binary.LittleEndian.PutUint32(buf, crc32.ChecksumIEEE(k.ED25519_PRI))
	k.CRC_ED25519_PRI = hex.EncodeToString(buf)
	binary.LittleEndian.PutUint32(buf, crc32.ChecksumIEEE(k.ED25519_PUB))
	k.CRC_ED25519_PUB = hex.EncodeToString(buf)
}

func (k *KeySet) LoadFromBytes(data []byte) error {
	return json.NewDecoder(bytes.NewReader(data)).Decode(k)
}

func (k *KeySet) SaveAsBytes() ([]byte, error) {
	buf, err := json.Marshal(k)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

var DefaultKeyGen KeyGen = GeneralKeyGen{}

type KeyGen interface {
	GenerateVitalKeySet() (*KeySet, error)

	Aes128() ([]byte, error)
}

type GeneralKeyGen struct {
}

func (g GeneralKeyGen) Aes128() ([]byte, error) {
	buf := make([]byte, 128/8)
	if n, err := rand.Read(buf); err != nil {
		return nil, err
	} else if n != len(buf) {
		return nil, fmt.Errorf("failed to read random data")
	}
	return buf, nil
}

func (g GeneralKeyGen) GenerateVitalKeySet() (*KeySet, error) {
	keySet := new(KeySet)
	{
		buf := make([]byte, 256/8)
		if n, err := rand.Read(buf); err != nil {
			return nil, err
		} else if n != len(buf) {
			return nil, fmt.Errorf("failed to read random data")
		}
		keySet.AES256 = buf
	}
	{
		pri, err := rsa.GenerateKey(rand.Reader, 4096)
		if err != nil {
			return nil, err
		}
		data := pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(pri),
		})
		keySet.RSA4096 = data
	}
	{
		pri, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
		if err != nil {
			return nil, err
		}
		keydata, err := x509.MarshalECPrivateKey(pri)
		if err != nil {
			return nil, err
		}
		data := pem.EncodeToMemory(&pem.Block{
			Type:  "EC PRIVATE KEY",
			Bytes: keydata,
		})
		keySet.ECDSA_P521 = data
	}
	{
		pri, pub, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return nil, err
		}
		keySet.ED25519_PUB = pub
		keySet.ED25519_PRI = pri
	}

	keySet.CalculateCrc()
	return keySet, nil
}
