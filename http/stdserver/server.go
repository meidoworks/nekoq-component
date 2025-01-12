package stdserver

import (
	"context"
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync/atomic"

	"github.com/meidoworks/nekoq-component/configure/secretapi"
)

type StdHttpServerReq struct {
	Addr    string
	Handler http.Handler

	StartedCallback func()
}

type StdHttpServer struct {
	l   net.Listener
	srv *http.Server
}

func StartStdHttpServer(req *StdHttpServerReq) (*StdHttpServer, error) {
	res := &StdHttpServer{}

	l, err := net.Listen("tcp", req.Addr)
	if err != nil {
		return nil, err
	}
	res.l = l
	res.srv = &http.Server{
		Handler: req.Handler,
	}

	go func() {
		if err := res.srv.Serve(l); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				panic(err)
			}
		}
	}()
	if req.StartedCallback != nil {
		req.StartedCallback()
	}

	return res, nil
}

func (h *StdHttpServer) Shutdown(ctx context.Context) error {
	return h.srv.Shutdown(ctx)
}

type StdHttpTlsServerReq struct {
	Addr    string
	Handler http.Handler

	Cert    *x509.Certificate
	CertKey crypto.PrivateKey

	StartedCallback func()
}

type certContainer struct {
	cert *x509.Certificate
	key  crypto.PrivateKey

	tlsCert *tls.Certificate
}

type StdHttpTlsServer struct {
	l   net.Listener
	srv *http.Server

	certContainer *atomic.Value
}

func (h *StdHttpTlsServer) UpdateCertificate(cert *x509.Certificate, key crypto.PrivateKey) error {
	// support dynamic updating certificate
	tlsCert, err := convertCertificate(cert, key)
	if err != nil {
		return err
	}
	h.certContainer.Store(&certContainer{
		cert:    cert,
		key:     key,
		tlsCert: tlsCert,
	})
	return nil
}

func StartStdHttpTlsServer(req *StdHttpTlsServerReq) (*StdHttpTlsServer, error) {
	result := &StdHttpTlsServer{
		certContainer: &atomic.Value{},
	}
	cert, err := convertCertificate(req.Cert, req.CertKey)
	if err != nil {
		return nil, err
	}
	result.certContainer.Store(&certContainer{
		cert:    req.Cert,
		key:     req.CertKey,
		tlsCert: cert,
	})

	ln, err := net.Listen("tcp", req.Addr)
	if err != nil {
		return nil, err
	}
	result.l = ln

	server := &http.Server{
		Handler: req.Handler,
		TLSConfig: &tls.Config{
			GetCertificate: func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
				cc := result.certContainer.Load().(*certContainer)
				if err := hello.SupportsCertificate(cc.tlsCert); err != nil {
					return nil, fmt.Errorf("unsupported certificate: %w", err)
				}
				return cc.tlsCert, nil
			},
		},
	}
	result.srv = server

	go func() {
		if err := result.srv.ServeTLS(ln, "", ""); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				panic(err)
			}
		}
	}()
	if req.StartedCallback != nil {
		req.StartedCallback()
	}

	return result, nil
}

func (h *StdHttpTlsServer) Shutdown(ctx context.Context) error {
	return h.srv.Shutdown(ctx)
}

func convertCertificate(cert *x509.Certificate, key crypto.PrivateKey) (*tls.Certificate, error) {
	certData, err := secretapi.NewPemTool().EncodeCertificate(cert)
	if err != nil {
		return nil, err
	}
	cert, err = secretapi.NewPemTool().ParseCertificate(certData)
	if err != nil {
		return nil, err
	}
	var tlsCert tls.Certificate
	tlsCert.Certificate = append(tlsCert.Certificate, cert.Raw)
	tlsCert.Leaf = cert
	tlsCert.PrivateKey = key
	return &tlsCert, nil
}

type CombinedStdHttpServerReq struct {
	Addr    string
	TlsAddr string
	Cert    *x509.Certificate
	CertKey crypto.PrivateKey
	Handler http.Handler

	StartedCallback func(serverTypeName string)
}

type CombinedStdHttpServer struct {
	httpServer    *StdHttpServer
	tlsHttpServer *StdHttpTlsServer
}

func StartCombinedStdHttpServer(req *CombinedStdHttpServerReq) (result *CombinedStdHttpServer, rerr error) {
	result = &CombinedStdHttpServer{}

	httpServer, err := StartStdHttpServer(&StdHttpServerReq{
		Addr:    req.Addr,
		Handler: req.Handler,
		StartedCallback: func() {
			if req.StartedCallback != nil {
				req.StartedCallback("http")
			}
		},
	})
	if err != nil {
		rerr = err
		return
	}
	result.httpServer = httpServer

	if req.TlsAddr != "" && req.Cert != nil && req.CertKey != nil {
		tlsHttpServer, err := StartStdHttpTlsServer(&StdHttpTlsServerReq{
			Addr:    req.TlsAddr,
			Handler: req.Handler,
			Cert:    req.Cert,
			CertKey: req.CertKey,
			StartedCallback: func() {
				if req.StartedCallback != nil {
					req.StartedCallback("https")
				}
			},
		})
		if err != nil {
			rerr = err
			return
		}
		result.tlsHttpServer = tlsHttpServer
	}
	return
}

func (h *CombinedStdHttpServer) Shutdown(ctx context.Context) error {
	errs := make([]error, 0)
	if h.httpServer != nil {
		if err := h.httpServer.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
	}
	if h.tlsHttpServer != nil {
		if err := h.tlsHttpServer.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}
