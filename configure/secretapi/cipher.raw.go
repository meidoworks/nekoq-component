package secretapi

import (
	"crypto"
	"errors"
)

type RawCipherTool struct {
}

func NewRawCipherTool() RawCipherTool {
	return RawCipherTool{}
}

func (r RawCipherTool) ConvertPKISupportedKey(keyType KeyType, key []byte) (crypto.PrivateKey, error) {
	switch keyType {
	case KeyRSA1024:
		fallthrough
	case KeyRSA2048:
		fallthrough
	case KeyRSA4096:
		fallthrough
	case KeyRSA3072:
		return NewPemTool().ParseRsaPrivateKey(key)
	case KeyECDSA224:
		fallthrough
	case KeyECDSA256:
		fallthrough
	case KeyECDSA384:
		fallthrough
	case KeyECDSA521:
		return NewPemTool().ParseECDSAPrivateKey(key)
	default:
		return nil, errors.New("ConvertPKISupportedKey: unsupported key type")
	}
}
