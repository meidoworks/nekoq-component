package secretserver

import (
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	"github.com/meidoworks/nekoq-component/configure/permissions"
	"github.com/meidoworks/nekoq-component/configure/secretaddon"
	"github.com/meidoworks/nekoq-component/configure/secretapi"
	"github.com/meidoworks/nekoq-component/http/chi2"
)

// JwtManageCreateAdminJwt generates supervisor jwt tokens
// This token only contains the permissions for:
//   - Next Admin Jwt Generation
//   - Creating New Jwt with other permissions
type JwtManageCreateAdminJwt struct {
	chi2.Controller
	Method string `method:"POST"`
	URL    string `url:"/api/v1/secret/jwt/admin/new"`

	signer   secretapi.JwtSigner
	verifier secretapi.JwtVerifier
}

type jwtAdminReq struct {
	Key string `json:"key"`
	Alg string `json:"alg"`
	TTL int    `json:"ttl"`
}

func (j *JwtManageCreateAdminJwt) HandleHttp(w http.ResponseWriter, r *http.Request) chi2.Render {
	obj, err := j.ParseBody(r)
	if err != nil {
		return chi2.NewErrRender(err)
	}
	req := obj.(*jwtAdminReq)
	tool := secretaddon.NewJwtTool(j.verifier)
	jwtData := secretapi.JwtData{}
	tool.SetupPermissions(jwtData, secretaddon.PermissionResourceList{}.
		Add(permissions.SecretJwtAdmin, permissions.SecretJwtNew))
	token, err := j.signer.SignJwt(req.Key, secretapi.JwtAlg(req.Alg), jwtData, secretapi.JwtOption{
		TTL: time.Duration(req.TTL) * time.Second,
	})
	if err != nil {
		return chi2.NewErrRender(err)
	}
	return chi2.NewJsonOkRender(map[string]interface{}{
		"token": token,
	})
}

func NewJwtManageCreateAdminJwt(signer secretapi.JwtSigner, verifier secretapi.JwtVerifier) *JwtManageCreateAdminJwt {
	return &JwtManageCreateAdminJwt{
		Controller: chi2.Controller{
			Middlewares: chi2.Middlewares{
				middleware.Logger,
				middleware.RealIP,
				middleware.Recoverer,
			},
			RequestValidators: chi2.RequestValidators{
				chi2.AllowContentTypeFor("application/json"),
				ValidateJwtToken(verifier, secretaddon.PermissionResourceList{}.
					Add(permissions.SecretJwtAdmin)),
			},
			BodyParser: func(hr *http.Request, r io.Reader) (any, error) {
				req := new(jwtAdminReq)
				if err := render.DecodeJSON(r, req); err != nil {
					return nil, err
				}
				return req, nil
			},
		},
		signer:   signer,
		verifier: verifier,
	}
}

// JwtManageCreateNewJwt generates general jwt tokens
// This token contains the permissions for:
//   - Specified permissions for general purpose except JwtAdmin permission
type JwtManageCreateNewJwt struct {
	chi2.Controller
	Method string `method:"POST"`
	URL    string `url:"/api/v1/secret/jwt/new"`

	signer   secretapi.JwtSigner
	verifier secretapi.JwtVerifier
}

type jwtNewReq struct {
	Key string `json:"key"`
	Alg string `json:"alg"`
	TTL int    `json:"ttl"`

	Permissions []string `json:"permissions"`
}

func (j *JwtManageCreateNewJwt) HandleHttp(w http.ResponseWriter, r *http.Request) chi2.Render {
	obj, err := j.ParseBody(r)
	if err != nil {
		return chi2.NewErrRender(err)
	}
	req := obj.(*jwtNewReq)
	if len(req.Permissions) == 0 || len(req.Permissions) > MaxPermissionsInSingleToken { // limit the count to reduce the size
		return chi2.NewStatusRender(http.StatusBadRequest)
	}

	// check and prepare permissions
	var perms []permissions.PermissionDef
	for _, perm := range req.Permissions {
		def, ok := permissions.GetPermissionDef(perm)
		if !ok {
			//FIXME should respond reason?
			return chi2.NewStatusRender(http.StatusBadRequest)
		}
		if def.Equals(permissions.SecretJwtAdmin) {
			return chi2.NewStatusRender(http.StatusUnauthorized)
		}
		perms = append(perms, def)
	}

	tool := secretaddon.NewJwtTool(j.verifier)
	jwtData := secretapi.JwtData{}
	tool.SetupPermissions(jwtData, secretaddon.PermissionResourceList{}.
		Add(perms...))
	token, err := j.signer.SignJwt(req.Key, secretapi.JwtAlg(req.Alg), jwtData, secretapi.JwtOption{
		TTL: time.Duration(req.TTL) * time.Second,
	})
	if err != nil {
		return chi2.NewErrRender(err)
	}
	return chi2.NewJsonOkRender(map[string]interface{}{
		"token": token,
	})
}

func NewJwtManageCreateNewJwt(signer secretapi.JwtSigner, verifier secretapi.JwtVerifier) *JwtManageCreateNewJwt {
	return &JwtManageCreateNewJwt{
		Controller: chi2.Controller{
			Middlewares: chi2.Middlewares{
				middleware.Logger,
				middleware.RealIP,
				middleware.Recoverer,
			},
			RequestValidators: chi2.RequestValidators{
				chi2.AllowContentTypeFor("application/json"),
				ValidateJwtToken(verifier, secretaddon.PermissionResourceList{}.
					Add(permissions.SecretJwtNew)),
			},
			BodyParser: func(hr *http.Request, r io.Reader) (any, error) {
				req := new(jwtNewReq)
				if err := render.DecodeJSON(r, req); err != nil {
					return nil, err
				}
				return req, nil
			},
		},
		signer:   signer,
		verifier: verifier,
	}
}

type JwtManageVerifyJwt struct {
	chi2.Controller
	Method string `method:"PUT"`
	URL    string `url:"/api/v1/secret/jwt/verify"`

	verifier secretapi.JwtVerifier
}

type jwtVerifyReq struct {
	Token string `json:"token"`
}

func (j *JwtManageVerifyJwt) HandleHttp(w http.ResponseWriter, r *http.Request) chi2.Render {
	obj, err := j.ParseBody(r)
	if err != nil {
		return chi2.NewErrRender(err)
	}
	req := obj.(*jwtVerifyReq)
	_, err = j.verifier.VerifyJwt(req.Token)
	if err != nil {
		//FIXME should log error?
		return chi2.NewJsonOkRender(map[string]interface{}{
			"result": false,
		})
	} else {
		return chi2.NewJsonOkRender(map[string]interface{}{
			"result": true,
		})
	}
}

func NewJwtManageVerifyJwt(verifier secretapi.JwtVerifier) *JwtManageVerifyJwt {
	return &JwtManageVerifyJwt{
		Controller: chi2.Controller{
			Middlewares: chi2.Middlewares{
				middleware.Logger,
				middleware.RealIP,
				middleware.Recoverer,
			},
			RequestValidators: chi2.RequestValidators{
				chi2.AllowContentTypeFor("application/json"),
				ValidateJwtToken(verifier, secretaddon.PermissionResourceList{}.
					Add(permissions.SecretJwtVerify)),
			},
			BodyParser: func(hr *http.Request, r io.Reader) (any, error) {
				req := new(jwtVerifyReq)
				if err := render.DecodeJSON(r, req); err != nil {
					return nil, err
				}
				return req, nil
			},
		},
		verifier: verifier,
	}
}
