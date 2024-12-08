package secretapi

import (
	"errors"
	"fmt"
)

var (
	ErrLevel1KeySetIsNotLoaded   = errors.New("secretapi: secret store does not have the level 1 key loaded")
	ErrNoStoredUnsealTokenExists = errors.New("secretapi: secret store does not have the unsealed token")
)

type KeyType int

// TODO optimize via mapping table
func (k KeyType) String() string {
	switch k {
	case KeyKeySet:
		return "KeySet"
	case KeyAES128:
		return "KeyAES128"
	case KeyAES192:
		return "KeyAES192"
	case KeyAES256:
		return "KeyAES256"
	case KeyEd25519:
		return "KeyEd25519"
	case KeyRSA1024:
		return "KeyRSA1024"
	case KeyRSA2048:
		return "KeyRSA2048"
	case KeyRSA4096:
		return "KeyRSA4096"
	case KeyRSA3072:
		return "KeyRSA3072"
	case KeyECDSA224:
		return "KeyECDSA224"
	case KeyECDSA256:
		return "KeyECDSA256"
	case KeyECDSA384:
		return "KeyECDSA384"
	case KeyECDSA521:
		return "KeyECDSA521"
	case KeyGeneral64B:
		return "KeyGeneral64B"
	case KeyGeneral128B:
		return "KeyGeneral128B"
	default:
		panic("unknown key type:" + fmt.Sprint(k))
	}
}

func (k *KeyType) FromString(str string) {
	switch str {
	case "KeySet":
		*k = KeyKeySet
	case "KeyAES128":
		*k = KeyAES128
	case "KeyAES192":
		*k = KeyAES192
	case "KeyAES256":
		*k = KeyAES256
	case "KeyEd25519":
		*k = KeyEd25519
	case "KeyRSA1024":
		*k = KeyRSA1024
	case "KeyRSA2048":
		*k = KeyRSA2048
	case "KeyRSA4096":
		*k = KeyRSA4096
	case "KeyRSA3072":
		*k = KeyRSA3072
	case "KeyECDSA224":
		*k = KeyECDSA224
	case "KeyECDSA256":
		*k = KeyECDSA256
	case "KeyECDSA384":
		*k = KeyECDSA384
	case "KeyECDSA521":
		*k = KeyECDSA521
	case "KeyGeneral64B":
		*k = KeyGeneral64B
	case "KeyGeneral128B":
		*k = KeyGeneral128B
	default:
		panic("unknown key type:" + str)
	}
}

const (
	KeyKeySet KeyType = 501
)

const (
	KeyAES128 KeyType = iota + 1001
	KeyAES192
	KeyAES256
)

const (
	KeyRSA1024 KeyType = iota + 2001
	KeyRSA2048
	KeyRSA3072
	KeyRSA4096
)

const (
	KeyECDSA224 KeyType = iota + 2101
	KeyECDSA256
	KeyECDSA384
	KeyECDSA521
)

const (
	KeyEd25519 KeyType = iota + 2201
)

const (
	KeyGeneral64B KeyType = iota + 3001
	KeyGeneral128B
)

type KeySetAlg int

const (
	KeySetAes     KeySetAlg = iota + 1 // only enc/dec
	KeySetRsa                          // both enc/dec and sign/verify
	KeySetECDSA                        // only sign/verify
	KeySetEd25519                      // only sign/verify
)

type KeyStorage interface {
	// SetupUnsealProviderAndWait sets up UnsealProvider for other operations which depend on it
	SetupUnsealProviderAndWait(provider UnsealProvider) error

	// StoreLevel1KeySet will create or rotate the level1 KeySet for the corresponding name with the given key
	StoreLevel1KeySet(name string, key *KeySet) error
	// StoreLevel2KeySet will create or rotate the level2 KeySet for the corresponding level1 and level2 names with the given key
	StoreLevel2KeySet(level1KeyName, name string, key *KeySet) error
	// FetchLevel2KeySet fetches the raw key data for external use
	FetchLevel2KeySet(name string) (int64, *KeySet, error)
	// StoreL2DataKey will create or rotate the level2 specific type of key
	StoreL2DataKey(l1KeyName, name string, keyType KeyType, key []byte) error
	// FetchL2DataKey fetches the raw data for external use
	FetchL2DataKey(name string) (int64, KeyType, []byte, error)

	// LoadLevel2KeySetById loads KeySet by id
	//
	// This method should be only used for decryption since it will retrieve the key regardless of key status.
	LoadLevel2KeySetById(id int64) (*KeySet, error)
	// LoadL2DataKeyById loads data key by id
	//
	// This method should be only used for decryption since it will retrieve the key regardless of key status.
	LoadL2DataKeyById(id int64) (KeyType, []byte, error)
}

type DefaultKeyStorage struct {
}

func (d *DefaultKeyStorage) SetupUnsealProviderAndWait(provider UnsealProvider) error {
	panic("unsupported operation")
}

func (d *DefaultKeyStorage) StoreLevel1KeySet(name string, key *KeySet) error {
	panic("unsupported operation")
}

func (d *DefaultKeyStorage) StoreLevel2KeySet(level1KeyName, name string, key *KeySet) error {
	panic("unsupported operation")
}

func (d *DefaultKeyStorage) FetchLevel2KeySet(name string) (int64, *KeySet, error) {
	panic("unsupported operation")
}

func (d *DefaultKeyStorage) StoreL2DataKey(l1KeyName, name string, keyType KeyType, key []byte) error {
	panic("unsupported operation")
}

func (d *DefaultKeyStorage) FetchL2DataKey(name string) (int64, KeyType, []byte, error) {
	panic("unsupported operation")
}

func (d *DefaultKeyStorage) LoadLevel2KeySetById(id int64) (*KeySet, error) {
	panic("unsupported operation")
}

func (d *DefaultKeyStorage) LoadL2DataKeyById(id int64) (KeyType, []byte, error) {
	panic("unsupported operation")
}
