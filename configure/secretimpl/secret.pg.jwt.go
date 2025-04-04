package secretimpl

import (
	"errors"
	"time"

	"github.com/meidoworks/nekoq-component/configure/secretaddon"
	"github.com/meidoworks/nekoq-component/configure/secretapi"
)

var (
	ErrJwtTokenInvalid = errors.New("jwt token invalid")
)

var _ secretapi.JwtSigner = new(PostgresKeyStorage)
var _ secretapi.JwtVerifier = new(PostgresKeyStorage)

func (p *PostgresKeyStorage) SignJwt(l2key string, jwtAlg secretapi.JwtAlg, data secretapi.JwtData, opt secretapi.JwtOption) (string, error) {
	tool := secretaddon.NewAddonTool(p)
	token, err := tool.SignJwtToken(l2key, jwtAlg, secretaddon.JwtClaims{}.FromJwtData(data).SetExp(time.Now().Add(opt.TTL)))
	if err != nil {
		return "", nil
	}

	if opt.ServerControl {
		//FIXME handle persistent in database
	}

	return token, nil
}

func (p *PostgresKeyStorage) VerifyJwt(jwtToken string) (secretapi.JwtData, error) {
	tool := secretaddon.NewAddonTool(p)
	claims, err := tool.VerifyJwtToken(jwtToken)
	if err != nil {
		return secretapi.JwtData{}, err
	}

	//FIXME handle check from persistent

	return claims, nil
}
