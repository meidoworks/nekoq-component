package secretserver

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	"github.com/meidoworks/nekoq-component/configure/permissions"
	"github.com/meidoworks/nekoq-component/configure/secretaddon"
	"github.com/meidoworks/nekoq-component/configure/secretapi"
	"github.com/meidoworks/nekoq-component/http/chi2"
)

// KeyManageGenerateNewKey generates L2 keys
type KeyManageGenerateNewKey struct {
	chi2.Controller
	Method string `method:"PUT"`
	URL    string `url:"/api/v1/secret/key/new/{key_name}"`

	keyStorage secretapi.KeyStorage
}

type newKeyReq struct {
	Type           string `json:"type"`             // RSA, ECDSA
	KeyScale       string `json:"key_scale"`        // RSA:1024,2048,3072,4096  ECDSA:224,256,384,521
	EncryptKeyName string `json:"encrypt_key_name"` // Should be Level1 key
}

func (k *KeyManageGenerateNewKey) HandleHttp(w http.ResponseWriter, r *http.Request) chi2.Render {
	obj, err := k.ParseBody(r)
	if err != nil {
		return chi2.NewErrRender(err)
	}
	req := obj.(*newKeyReq)
	keyName := strings.TrimSpace(chi.URLParam(r, "key_name"))
	encryptKeyName := strings.TrimSpace(req.EncryptKeyName)
	if keyName == "" || encryptKeyName == "" {
		return chi2.NewStatusRender(http.StatusBadRequest)
	}
	switch req.Type {
	case "RSA":
		switch req.KeyScale {
		case "2048", "1024", "3072", "4096":
		default:
			return chi2.NewStatusRender(http.StatusBadRequest)
		}
	case "ECDSA":
		switch req.KeyScale {
		case "224", "256", "384", "521":
		default:
			return chi2.NewStatusRender(http.StatusBadRequest)
		}
	default:
		return chi2.NewStatusRender(http.StatusBadRequest)
	}

	var key []byte
	var kt secretapi.KeyType
	switch req.Type {
	case "RSA":
		var keyType secretapi.KeyType
		switch req.KeyScale {
		case "1024":
			keyType = secretapi.KeyRSA1024
		case "2048":
			keyType = secretapi.KeyRSA2048
		case "3072":
			keyType = secretapi.KeyRSA3072
		case "4096":
			keyType = secretapi.KeyRSA4096
		default:
			panic(errors.New("unsupported key type"))
		}
		if k, err := secretapi.DefaultKeyGen.RSA(keyType); err != nil {
			return chi2.NewErrRender(err)
		} else {
			key = k
			kt = keyType
		}
	case "ECDSA":
		var keyType secretapi.KeyType
		switch req.KeyScale {
		case "256":
			keyType = secretapi.KeyECDSA256
		case "224":
			keyType = secretapi.KeyECDSA224
		case "384":
			keyType = secretapi.KeyECDSA384
		case "521":
			keyType = secretapi.KeyECDSA521
		default:
			panic(errors.New("unsupported key type"))
		}
		if k, err := secretapi.DefaultKeyGen.ECDSA(keyType); err != nil {
			return chi2.NewErrRender(err)
		} else {
			key = k
			kt = keyType
		}
	default:
		panic(errors.New("unsupported key type"))
	}

	if err := k.keyStorage.StoreL2DataKey(encryptKeyName, keyName, kt, key); err != nil {
		return chi2.NewErrRender(err)
	}

	return chi2.NewJsonOkRender(map[string]interface{}{
		"key_name": keyName,
		//FIXME perhaps need respond key_version
		//"key_version": 0,
	})
}

func NewKeyManageGenerateNewKey(verifier secretapi.JwtVerifier, keyStorage secretapi.KeyStorage) *KeyManageGenerateNewKey {
	return &KeyManageGenerateNewKey{
		Controller: chi2.Controller{
			Middlewares: chi2.Middlewares{
				middleware.Logger,
				middleware.RealIP,
				middleware.Recoverer,
			},
			RequestValidators: chi2.RequestValidators{
				chi2.AllowContentTypeFor("application/json"),
				ValidateJwtToken(verifier, secretaddon.PermissionResourceList{}.
					Add(permissions.SecretKeyAdmin)),
			},
			BodyParser: func(hr *http.Request, r io.Reader) (any, error) {
				req := new(newKeyReq)
				if err := render.DecodeJSON(r, req); err != nil {
					return nil, err
				}
				return req, nil
			},
		},
		keyStorage: keyStorage,
	}
}

// KeyManageGetKeyById retrieves level2 key by key id
type KeyManageGetKeyById struct {
	chi2.Controller
	Method string `method:"GET"`
	URL    string `url:"/api/v1/secret/key/info/{key_id}"`

	keyStorage secretapi.KeyStorage
}

func (k *KeyManageGetKeyById) HandleHttp(w http.ResponseWriter, r *http.Request) chi2.Render {
	keyId := strings.TrimSpace(chi.URLParam(r, "key_id"))
	if keyId == "" {
		return chi2.NewStatusRender(http.StatusBadRequest)
	}
	keyIdInt, err := strconv.ParseInt(keyId, 10, 64)
	if err != nil {
		return chi2.NewErrRender(err)
	}
	kt, key, err := k.keyStorage.LoadL2DataKeyById(keyIdInt)
	if err != nil {
		return chi2.NewErrRender(err)
	}
	var keyTypeSeries string
	var format = "unknown"
	switch kt {
	case secretapi.KeyKeySet:
		keyTypeSeries = "keyset"
		format = "pem"
	case secretapi.KeyAES128, secretapi.KeyAES192, secretapi.KeyAES256:
		keyTypeSeries = "aes"
		format = "raw"
	case secretapi.KeyRSA1024, secretapi.KeyRSA2048, secretapi.KeyRSA3072, secretapi.KeyRSA4096:
		keyTypeSeries = "rsa"
		format = "pem"
	case secretapi.KeyECDSA224, secretapi.KeyECDSA256, secretapi.KeyECDSA384, secretapi.KeyECDSA521:
		keyTypeSeries = "ecdsa"
		format = "pem"
	case secretapi.KeyEd25519:
		keyTypeSeries = "ed25519" //FIXME note: ed25519 key may not be independently stored
		format = "raw"
	case secretapi.KeyGeneral64B, secretapi.KeyGeneral128B:
		keyTypeSeries = "general"
		format = "raw"
	default:
		return chi2.NewErrRender(errors.New("unsupported key type"))
	}

	return chi2.NewJsonOkRender(map[string]interface{}{
		"key_type": keyTypeSeries,
		"key":      key,
		"format":   format,
	})
}

func NewKeyManageGetKeyById(verifier secretapi.JwtVerifier, keyStorage secretapi.KeyStorage) *KeyManageGetKeyById {
	return &KeyManageGetKeyById{
		Controller: chi2.Controller{
			Middlewares: chi2.Middlewares{
				middleware.Logger,
				middleware.RealIP,
				middleware.Recoverer,
			},
			RequestValidators: chi2.RequestValidators{
				ValidateJwtToken(verifier, secretaddon.PermissionResourceList{}.
					Add(permissions.SecretKeyAdmin)),
			},
			BodyParser: func(hr *http.Request, r io.Reader) (any, error) {
				req := new(newKeyReq)
				if err := render.DecodeJSON(r, req); err != nil {
					return nil, err
				}
				return req, nil
			},
		},
		keyStorage: keyStorage,
	}
}
