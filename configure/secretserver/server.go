package secretserver

import (
	"context"
	"crypto"
	"crypto/x509"
	"errors"
	"log"

	"github.com/go-chi/chi/v5"

	"github.com/meidoworks/nekoq-component/configure/secretapi"
	"github.com/meidoworks/nekoq-component/http/chi2"
	"github.com/meidoworks/nekoq-component/http/stdserver"
)

type SecretServerReq struct {
	ReadService struct {
		Addr    string
		TlsAddr string
		Cert    *x509.Certificate
		CertKey crypto.PrivateKey
	}
	WriteService struct {
		Addr    string
		TlsAddr string
		Cert    *x509.Certificate
		CertKey crypto.PrivateKey
	}

	JwtSigner   secretapi.JwtSigner
	JwtVerifier secretapi.JwtVerifier
	KeyStorage  secretapi.KeyStorage
	CertStorage secretapi.CertStorage

	DebugOpt struct {
		PrintRegisteredAPIs bool
	}
}

type SecretServer struct {
	req      *SecretServerReq
	readMux  *chi.Mux
	writeMux *chi.Mux

	writeserver *stdserver.CombinedStdHttpServer
	readserver  *stdserver.CombinedStdHttpServer
}

func (s *SecretServer) initReadMux() {
	r := chi.NewRouter()

	s.readMux = r
}

func (s *SecretServer) initWriteMux() error {
	r := chi.NewRouter()

	if err := initWriteApis(r, s.req); err != nil {
		return err
	}

	s.writeMux = r
	return nil
}

func initWriteApis(r *chi.Mux, req *SecretServerReq) error {
	stub := chi2.NewChiApiStub()
	if err := stub.RegisterControllers(
		NewJwtManageCreateAdminJwt(req.JwtSigner, req.JwtVerifier),
		NewJwtManageVerifyJwt(req.JwtVerifier),
		NewJwtManageCreateNewJwt(req.JwtSigner, req.JwtVerifier)); err != nil {
		return err
	}
	if err := stub.RegisterControllers(
		NewCertManageCreateCertReq(req.JwtVerifier, req.KeyStorage),
		NewCertManageCreateCert(req.JwtVerifier, req.KeyStorage, req.CertStorage),
		NewCertManageGetCert(req.JwtVerifier, req.KeyStorage, req.CertStorage)); err != nil {
		return err
	}
	if err := stub.RegisterControllers(
		NewKeyManageGenerateNewKey(req.JwtVerifier, req.KeyStorage),
		NewKeyManageGetKeyById(req.JwtVerifier, req.KeyStorage),
		NewKeyManageGetKeyByName(req.JwtVerifier, req.KeyStorage)); err != nil {
		return err
	}
	if req.DebugOpt.PrintRegisteredAPIs {
		stub.LogAllControllers()
	}
	stub.BuildFor(r)

	return nil
}

func (s *SecretServer) Startup() error {
	writeserver, err := stdserver.StartCombinedStdHttpServer(&stdserver.CombinedStdHttpServerReq{
		Addr:    s.req.WriteService.Addr,
		TlsAddr: s.req.WriteService.TlsAddr,
		Cert:    s.req.WriteService.Cert,
		CertKey: s.req.WriteService.CertKey,
		Handler: s.writeMux,
		StartedCallback: func(serverTypeName string) {
			log.Println("SecretServer [" + serverTypeName + "] write endpoint started.")
		},
	})
	if err != nil {
		return err
	}
	s.writeserver = writeserver
	readserver, err := stdserver.StartCombinedStdHttpServer(&stdserver.CombinedStdHttpServerReq{
		Addr:    s.req.ReadService.Addr,
		TlsAddr: s.req.ReadService.TlsAddr,
		Cert:    s.req.ReadService.Cert,
		CertKey: s.req.ReadService.CertKey,
		Handler: s.readMux,
		StartedCallback: func(serverTypeName string) {
			log.Println("SecretServer [" + serverTypeName + "] endpoint started.")
		},
	})
	if err != nil {
		return err
	}
	s.readserver = readserver
	return nil
}

func (s *SecretServer) Shutdown(ctx context.Context) error {
	var errs []error
	if s.readserver != nil {
		if err := s.readserver.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
	}
	if s.writeserver != nil {
		if err := s.writeserver.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	} else {
		return nil
	}
}

func NewSecretServer(req *SecretServerReq) (*SecretServer, error) {
	server := &SecretServer{
		req: req,
	}

	server.initReadMux()
	if err := server.initWriteMux(); err != nil {
		return nil, err
	}

	return server, nil
}
