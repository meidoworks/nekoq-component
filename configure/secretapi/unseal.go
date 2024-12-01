package secretapi

import "context"

type UnsealResponse struct {
	Selector string
	Token    string
}

type UnsealProvider interface {
	// WaitUnsealOperation waits for provider to finish unseal operation.
	// The provider may trigger external flows to allow users to provide unseal information.
	// The caller may cancel the operation once it reaches the deadline via the context.
	// The provider should kick off the previous operation if a new one triggered in order to ensure the security.
	WaitUnsealOperation(ctx context.Context, encToken string) (*UnsealResponse, error)

	UseKeyId() string

	Encrypt(ctx context.Context, plaintext []byte) ([]byte, string, error)
	Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error)
}
