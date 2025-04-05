package secretapi

import "time"

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

type JwtData map[string]any

type JwtOption struct {
	// Is the JWT token controlled by the server.
	// Checks include revoke, expiration, etc.
	ServerControl bool
	// Time-to-live
	TTL time.Duration
	// One time JWT token
	OneTime bool
}

type JwtSigner interface {
	SignJwt(l2key string, jwtAlg JwtAlg, data JwtData, opt JwtOption) (string, error)
}

type JwtVerifier interface {
	VerifyJwt(jwtToken string) (JwtData, error)
}
