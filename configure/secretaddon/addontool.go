package secretaddon

import (
	"errors"
	"strconv"
	"time"

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

type JwtAlg = secretapi.JwtAlg

type JwtClaims map[string]interface{}

func (j JwtClaims) SetIssuer(issuer string) JwtClaims {
	j["iss"] = issuer
	return j
}

func (j JwtClaims) SetSubject(subject string) JwtClaims {
	j["sub"] = subject
	return j
}

func (j JwtClaims) SetAudience(audiences []string) JwtClaims {
	j["aud"] = audiences
	return j
}

func (j JwtClaims) SetExp(t time.Time) JwtClaims {
	j["exp"] = t.Truncate(time.Second).Unix()
	return j
}

func (j JwtClaims) SetNbf(t time.Time) JwtClaims {
	j["nbf"] = t.Truncate(time.Second).Unix()
	return j
}

func (j JwtClaims) SetIat(t time.Time) JwtClaims {
	j["iat"] = t.Truncate(time.Second).Unix()
	return j
}

func (j JwtClaims) SetID(id string) JwtClaims {
	j["jti"] = id
	return j
}

func (j JwtClaims) FromJwtData(data secretapi.JwtData) JwtClaims {
	for key, value := range data {
		j[key] = value
	}
	return j
}

type AddonTool struct {
	keyStorage secretapi.KeyStorage
}

func NewAddonTool(keyStorage secretapi.KeyStorage) *AddonTool {
	return &AddonTool{
		keyStorage: keyStorage,
	}
}

func (a *AddonTool) SignJwtToken(keyName string, jwtAlg JwtAlg, claims JwtClaims) (string, error) {
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

func convertClaims(claims JwtClaims) jwt.MapClaims {
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
		if alg != secretapi.JwtAlgHS256 {
			return nil, nil, errors.New("invalid algorithm")
		}
		return jwt.SigningMethodHS256, key, nil
	case secretapi.KeyGeneral128B:
		if alg == secretapi.JwtAlgHS384 {
			return jwt.SigningMethodHS384, key, nil
		} else if alg == secretapi.JwtAlgHS512 {
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
		case secretapi.JwtAlgRS256:
			return jwt.SigningMethodRS256, priKey.Public(), nil
		case secretapi.JwtAlgRS384:
			return jwt.SigningMethodRS384, priKey.Public(), nil
		case secretapi.JwtAlgRS512:
			return jwt.SigningMethodRS512, priKey.Public(), nil
		case secretapi.JwtAlgPS256:
			return jwt.SigningMethodPS256, priKey.Public(), nil
		case secretapi.JwtAlgPS384:
			return jwt.SigningMethodPS384, priKey.Public(), nil
		case secretapi.JwtAlgPS512:
			if keyType == secretapi.KeyRSA1024 {
				return nil, nil, ErrJwtInvalidAlgCombination
			}
			return jwt.SigningMethodPS512, priKey.Public(), nil
		default:
			return nil, nil, errors.New("invalid algorithm")
		}
	case secretapi.KeyECDSA256:
		if alg != secretapi.JwtAlgES256 {
			return nil, nil, errors.New("invalid key type")
		}
		priKey, err := secretapi.NewPemTool().ParseECDSAPrivateKey(key)
		if err != nil {
			return nil, nil, err
		}
		return jwt.SigningMethodES256, priKey.Public(), nil
	case secretapi.KeyECDSA384:
		if alg != secretapi.JwtAlgES384 {
			return nil, nil, errors.New("invalid key type")
		}
		priKey, err := secretapi.NewPemTool().ParseECDSAPrivateKey(key)
		if err != nil {
			return nil, nil, err
		}
		return jwt.SigningMethodES384, priKey.Public(), nil
	case secretapi.KeyECDSA521:
		if alg != secretapi.JwtAlgES512 {
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
		if alg != secretapi.JwtAlgHS256 {
			return nil, nil, errors.New("invalid key type")
		}
		return jwt.SigningMethodHS256, key, nil
	case secretapi.KeyGeneral128B:
		if alg == secretapi.JwtAlgHS512 {
			return jwt.SigningMethodHS512, key, nil
		} else if alg == secretapi.JwtAlgHS384 {
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
		case secretapi.JwtAlgRS256:
			return jwt.SigningMethodRS256, priKey, nil
		case secretapi.JwtAlgRS384:
			return jwt.SigningMethodRS384, priKey, nil
		case secretapi.JwtAlgRS512:
			return jwt.SigningMethodRS512, priKey, nil
		case secretapi.JwtAlgPS256:
			return jwt.SigningMethodPS256, priKey, nil
		case secretapi.JwtAlgPS384:
			return jwt.SigningMethodPS384, priKey, nil
		case secretapi.JwtAlgPS512:
			if keyType == secretapi.KeyRSA1024 {
				return nil, nil, ErrJwtInvalidAlgCombination
			}
			return jwt.SigningMethodPS512, priKey, nil
		default:
			return nil, nil, errors.New("invalid algorithm")
		}
	case secretapi.KeyECDSA256:
		if alg != secretapi.JwtAlgES256 {
			return nil, nil, errors.New("invalid key type")
		}
		priKey, err := secretapi.NewPemTool().ParseECDSAPrivateKey(key)
		if err != nil {
			return nil, nil, err
		}
		return jwt.SigningMethodES256, priKey, nil
	case secretapi.KeyECDSA384:
		if alg != secretapi.JwtAlgES384 {
			return nil, nil, errors.New("invalid key type")
		}
		priKey, err := secretapi.NewPemTool().ParseECDSAPrivateKey(key)
		if err != nil {
			return nil, nil, err
		}
		return jwt.SigningMethodES384, priKey, nil
	case secretapi.KeyECDSA521:
		if alg != secretapi.JwtAlgES512 {
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
