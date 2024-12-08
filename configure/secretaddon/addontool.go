package secretaddon

import (
	"errors"
	"strconv"

	"github.com/golang-jwt/jwt/v5"

	"github.com/meidoworks/nekoq-component/configure/secretapi"
)

var (
	ErrJwtInvalidAlgCombination = errors.New("invalid algorithm combination")
)

const (
	JwtHeaderKid = "kid"
	JwtHeaderAlg = "alg"
)

type JwtAlg string

const (
	JwtAlgHS256 JwtAlg = "HS256"
	JwtAlgHS384 JwtAlg = "HS384"
	JwtAlgHS512 JwtAlg = "HS512"
	JwtAlgRS256 JwtAlg = "RS256"
	JwtAlgRS384 JwtAlg = "RS384"
	JwtAlgRS512 JwtAlg = "RS512"
	JwtAlgPS256 JwtAlg = "PS256"
	JwtAlgPS384 JwtAlg = "PS384"
	JwtAlgPS512 JwtAlg = "PS512"
	JwtAlgES256 JwtAlg = "ES256"
	JwtAlgES384 JwtAlg = "ES384"
	JwtAlgES512 JwtAlg = "ES512"
)

type AddonTool struct {
	keyStorage secretapi.KeyStorage
}

func NewAddonTool(keyStorage secretapi.KeyStorage) *AddonTool {
	return &AddonTool{
		keyStorage: keyStorage,
	}
}

func (a *AddonTool) SignJwtToken(keyName string, jwtAlg JwtAlg, claims map[string]interface{}) (string, error) {
	keyId, keyType, key, err := a.keyStorage.FetchL2DataKey(keyName)
	if err != nil {
		return "", err
	}

	signingMethod, signingKey, err := jwtSigningKeyMapping(keyType, jwtAlg, key)
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(signingMethod, convertClaims(claims))
	token.Header[JwtHeaderKid] = keyIdString(keyId)

	return token.SignedString(signingKey)
}

func (a *AddonTool) VerifyJwtToken(tokenString string) (map[string]interface{}, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		alg, ok := token.Header[JwtHeaderAlg].(string)
		if !ok {
			return nil, errors.New("invalid algorithm")
		}
		keyIdStr, ok := token.Header[JwtHeaderKid].(string)
		if !ok {
			return nil, errors.New("kid is not a string")
		}
		keyId, err := keyIdVal(keyIdStr)
		if err != nil {
			return nil, err
		}

		kt, key, err := a.keyStorage.LoadL2DataKeyById(keyId)
		if err != nil {
			return nil, err
		}
		_, verifyKey, err := jwtVerificationKeyMapping(kt, JwtAlg(alg), key)
		if err != nil {
			return nil, err
		}

		return verifyKey, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

func convertClaims(claims map[string]interface{}) jwt.MapClaims {
	m := jwt.MapClaims{}
	for k, v := range claims {
		m[k] = v
	}
	return m
}

func jwtVerificationKeyMapping(keyType secretapi.KeyType, alg JwtAlg, key []byte) (jwt.SigningMethod, any, error) {
	//FIXME perhaps we have to change function return type

	// Automatically determine jwt alg using internal KeyType and convert key to specific type
	switch keyType {
	case secretapi.KeyGeneral64B:
		if alg != JwtAlgHS256 {
			return nil, nil, errors.New("invalid algorithm")
		}
		return jwt.SigningMethodHS256, key, nil
	case secretapi.KeyGeneral128B:
		if alg == JwtAlgHS384 {
			return jwt.SigningMethodHS384, key, nil
		} else if alg == JwtAlgHS512 {
			return jwt.SigningMethodHS512, key, nil
		}
		return nil, nil, errors.New("invalid algorithm")
	case secretapi.KeyRSA1024:
		fallthrough
	case secretapi.KeyRSA2048:
		fallthrough
	case secretapi.KeyRSA4096:
		fallthrough
	case secretapi.KeyRSA3072:
		priKey, err := secretapi.NewPemTool().ParseRsaPrivateKey(key)
		if err != nil {
			return nil, nil, err
		}
		switch alg {
		case JwtAlgRS256:
			return jwt.SigningMethodRS256, priKey.Public(), nil
		case JwtAlgRS384:
			return jwt.SigningMethodRS384, priKey.Public(), nil
		case JwtAlgRS512:
			return jwt.SigningMethodRS512, priKey.Public(), nil
		case JwtAlgPS256:
			return jwt.SigningMethodPS256, priKey.Public(), nil
		case JwtAlgPS384:
			return jwt.SigningMethodPS384, priKey.Public(), nil
		case JwtAlgPS512:
			if keyType == secretapi.KeyRSA1024 {
				return nil, nil, ErrJwtInvalidAlgCombination
			}
			return jwt.SigningMethodPS512, priKey.Public(), nil
		default:
			return nil, nil, errors.New("invalid algorithm")
		}
	case secretapi.KeyECDSA256:
		if alg != JwtAlgES256 {
			return nil, nil, errors.New("invalid key type")
		}
		priKey, err := secretapi.NewPemTool().ParseECDSAPrivateKey(key)
		if err != nil {
			return nil, nil, err
		}
		return jwt.SigningMethodES256, priKey.Public(), nil
	case secretapi.KeyECDSA384:
		if alg != JwtAlgES384 {
			return nil, nil, errors.New("invalid key type")
		}
		priKey, err := secretapi.NewPemTool().ParseECDSAPrivateKey(key)
		if err != nil {
			return nil, nil, err
		}
		return jwt.SigningMethodES384, priKey.Public(), nil
	case secretapi.KeyECDSA521:
		if alg != JwtAlgES512 {
			return nil, nil, errors.New("invalid key type")
		}
		priKey, err := secretapi.NewPemTool().ParseECDSAPrivateKey(key)
		if err != nil {
			return nil, nil, err
		}
		return jwt.SigningMethodES512, priKey.Public(), nil
	}

	return nil, nil, errors.New("invalid key type")
}

func jwtSigningKeyMapping(keyType secretapi.KeyType, alg JwtAlg, key []byte) (jwt.SigningMethod, any, error) {
	// Automatically determine jwt alg using internal KeyType and convert key to specific type
	switch keyType {
	case secretapi.KeyGeneral64B:
		if alg != JwtAlgHS256 {
			return nil, nil, errors.New("invalid key type")
		}
		return jwt.SigningMethodHS256, key, nil
	case secretapi.KeyGeneral128B:
		if alg == JwtAlgHS512 {
			return jwt.SigningMethodHS512, key, nil
		} else if alg == JwtAlgHS384 {
			return jwt.SigningMethodHS384, key, nil
		}
		return nil, nil, errors.New("invalid key type")
	case secretapi.KeyRSA1024:
		fallthrough
	case secretapi.KeyRSA2048:
		fallthrough
	case secretapi.KeyRSA4096:
		fallthrough
	case secretapi.KeyRSA3072:
		priKey, err := secretapi.NewPemTool().ParseRsaPrivateKey(key)
		if err != nil {
			return nil, nil, err
		}
		switch alg {
		case JwtAlgRS256:
			return jwt.SigningMethodRS256, priKey, nil
		case JwtAlgRS384:
			return jwt.SigningMethodRS384, priKey, nil
		case JwtAlgRS512:
			return jwt.SigningMethodRS512, priKey, nil
		case JwtAlgPS256:
			return jwt.SigningMethodPS256, priKey, nil
		case JwtAlgPS384:
			return jwt.SigningMethodPS384, priKey, nil
		case JwtAlgPS512:
			if keyType == secretapi.KeyRSA1024 {
				return nil, nil, ErrJwtInvalidAlgCombination
			}
			return jwt.SigningMethodPS512, priKey, nil
		default:
			return nil, nil, errors.New("invalid algorithm")
		}
	case secretapi.KeyECDSA256:
		if alg != JwtAlgES256 {
			return nil, nil, errors.New("invalid key type")
		}
		priKey, err := secretapi.NewPemTool().ParseECDSAPrivateKey(key)
		if err != nil {
			return nil, nil, err
		}
		return jwt.SigningMethodES256, priKey, nil
	case secretapi.KeyECDSA384:
		if alg != JwtAlgES384 {
			return nil, nil, errors.New("invalid key type")
		}
		priKey, err := secretapi.NewPemTool().ParseECDSAPrivateKey(key)
		if err != nil {
			return nil, nil, err
		}
		return jwt.SigningMethodES384, priKey, nil
	case secretapi.KeyECDSA521:
		if alg != JwtAlgES512 {
			return nil, nil, errors.New("invalid key type")
		}
		priKey, err := secretapi.NewPemTool().ParseECDSAPrivateKey(key)
		if err != nil {
			return nil, nil, err
		}
		return jwt.SigningMethodES512, priKey, nil
	}

	return nil, nil, errors.New("invalid key type")
}

func keyIdString(keyId int64) string {
	return strconv.FormatInt(keyId, 10)
}

func keyIdVal(keyIdStr string) (int64, error) {
	keyId, err := strconv.ParseInt(keyIdStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return keyId, nil
}
