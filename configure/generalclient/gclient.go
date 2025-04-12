package generalclient

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"time"
)

type ClientOptions struct {
	Secret struct {
		AddrList   []string
		Token      string
		EnforceTls bool
		ClientCert ClientCertOptions
		CaCert     CaCertOptions
	}
}

type ClientCertOptions struct {
	CertPath string
	KeyPath  string

	certData []byte
	keyData  []byte
}

type CaCertOptions struct {
	CertPath string

	certData []byte
}

func (c *ClientOptions) readFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)
	return io.ReadAll(f)
}

func (c *ClientOptions) preload() error {
	// load certs and keys
	{
		if c.Secret.CaCert.CertPath != "" {
			data, err := c.readFile(c.Secret.CaCert.CertPath)
			if err != nil {
				return err
			}
			c.Secret.CaCert.certData = data
		}
		if c.Secret.ClientCert.CertPath != "" {
			data, err := c.readFile(c.Secret.ClientCert.CertPath)
			if err != nil {
				return err
			}
			c.Secret.ClientCert.certData = data
		}
		if c.Secret.ClientCert.KeyPath != "" {
			data, err := c.readFile(c.Secret.ClientCert.KeyPath)
			if err != nil {
				return err
			}
			c.Secret.ClientCert.keyData = data
		}
	}
	// validate enforce https
	if c.Secret.EnforceTls {
		for _, addr := range c.Secret.AddrList {
			parsedURL, err := url.Parse(addr)
			if err != nil {
				return fmt.Errorf("invalid base URL: %v", err)
			}
			if parsedURL.Scheme != "https" {
				return fmt.Errorf("HTTP is not allowed; please use HTTPS")
			}
		}
	}
	return nil
}

type GeneralClient struct {
	opt *ClientOptions

	secretHttpClient *http.Client
}

func NewGeneralClient(opt *ClientOptions) (*GeneralClient, error) {
	if err := opt.preload(); err != nil {
		return nil, err
	}

	var secretHttpClient *http.Client
	{
		var tlsConfig *tls.Config
		if len(opt.Secret.CaCert.certData) > 0 {
			caCertPool := x509.NewCertPool()
			if ok := caCertPool.AppendCertsFromPEM(opt.Secret.CaCert.certData); !ok {
				return nil, fmt.Errorf("failed to append CA certificate")
			}
			tlsConfig = &tls.Config{
				RootCAs: caCertPool,
			}
		}
		transport := &http.Transport{
			MaxIdleConns:        16,
			MaxIdleConnsPerHost: 8,
			IdleConnTimeout:     30 * time.Second, // Timeout for idle connections
			TLSClientConfig:     tlsConfig,
		}
		secretHttpClient = &http.Client{Transport: transport}
	}

	return &GeneralClient{
		opt:              opt,
		secretHttpClient: secretHttpClient,
	}, nil
}

func (g *GeneralClient) randomAddr(list []string) string {
	return list[rand.Intn(len(list))]
}
