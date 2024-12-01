package secretapi

import "errors"

const (
	DefaultTokenString = "secret/default_token"

	TokenName = "secret_token"
)

var (
	ErrUnsealFailedOnMismatchToken = errors.New("unseal failed on mismatch token")
	ErrNoSuchKey                   = errors.New("no such key")
)
