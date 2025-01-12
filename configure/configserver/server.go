package configserver

import (
	"context"
	"crypto"
	"crypto/x509"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/meidoworks/nekoq-component/configure/configapi"
	"github.com/meidoworks/nekoq-component/http/stdserver"
)

type ConfigureOptions struct {
	Addr      string
	TLSConfig struct {
		Addr string
		Cert *x509.Certificate
		Key  crypto.PrivateKey
	}
	WriteApi struct {
		DataWriter configapi.DataWriter
		Addr       string
		TLSConfig  struct {
			Addr string
			Cert *x509.Certificate
			Key  crypto.PrivateKey
		}
	}

	MaxWaitTimeForUpdate int // in seconds

	DataPump          configapi.DataPump
	VersionComparator configapi.VersionComparator
}

func (c *ConfigureOptions) GetMaxWaitTimeForUpdate() time.Duration {
	if c.MaxWaitTimeForUpdate <= 0 {
		return 60 * time.Second
	} else {
		return time.Duration(c.MaxWaitTimeForUpdate) * time.Second
	}
}

type ConfigureServer struct {
	readMux *chi.Mux // for client read

	server      *server
	opt         ConfigureOptions
	writeServer struct {
		writeServer *writeServer
		writeMux    *chi.Mux // for management write
		httpServer  *stdserver.CombinedStdHttpServer
	}

	httpServer *stdserver.CombinedStdHttpServer
}

func (c *ConfigureServer) logError(messsage string, err error) {
	if err != nil {
		log.Println("[ERROR]", messsage, err)
	} else {
		log.Println("[ERROR]", messsage)
	}
}

func (c *ConfigureServer) logWarn(message string, args ...any) {
	log.Println(append([]any(nil), "[WARN]", message, args))
}

func (c *ConfigureServer) handleRetrieveAndListen(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Accept") != "application/cbor" {
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		c.logError("read http body error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	req := new(configapi.AcquireConfigurationReq)
	if err := cbor.Unmarshal(data, req); err != nil {
		c.logError("parse http body error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(req.Requested) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ch, cancelFn, err := c.server.RetrieveOrWait(req)
	if errors.Is(err, ErrHasUnknownConfiguration) {
		c.logError("some of the configuration not found", err)
		w.WriteHeader(http.StatusNotFound)
		//FIXME add details of missing configurations according to the api spec
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		//FIXME more details according to the api spec
		return
	}

	timer := time.NewTimer(c.opt.GetMaxWaitTimeForUpdate())
	var accumulated = make([]configapi.Configuration, 0, len(req.Requested))
	defer timer.Stop()
	defer cancelFn()
	select {
	case obj, ok := <-ch:
		for {
			if !ok {
				if len(accumulated) == 0 {
					c.logError("wait result should not be empty", nil)
					w.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					obj := &configapi.AcquireConfigurationRes{
						Requested: accumulated,
					}
					if data, err := cbor.Marshal(obj); err != nil {
						c.logError("marshal result failed", err)
						w.WriteHeader(http.StatusInternalServerError)
						return
					} else {
						w.Header().Add("Content-Type", "application/cbor")
						w.WriteHeader(http.StatusOK)
						if _, err := w.Write(data); err != nil {
							c.logError("write http body failed", err)
							return
						} else {
							return
						}
					}
				}
			}
			accumulated = append(accumulated, *obj.Configuration)
			obj, ok = <-ch // no need to check timer since any update will cause the channel to be closed
		}
	case <-timer.C:
		w.WriteHeader(http.StatusNotModified)
		return
	}
}

func (c *ConfigureServer) handleGetConfiguration(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Accept") != "application/cbor" {
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}
	selectorsInfo := strings.TrimSpace(r.Header.Get("X-Configuration-Sel"))
	if selectorsInfo == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	optSelectorsInfo := strings.TrimSpace(r.Header.Get("X-Configuration-Opt-Sel"))

	group := chi.URLParam(r, "group")
	if group == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	key := chi.URLParam(r, "key")
	if key == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	cfg, err := c.server.GetConfigurationViaPlainRequest(group, key, selectorsInfo, optSelectorsInfo)
	if errors.Is(err, ErrHasUnknownConfiguration) {
		c.logError("the configuration not found", err)
		//FIXME add details of missing configurations according to the api spec
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		//FIXME more details according to the api spec
		return
	}

	obj := &configapi.GetConfigurationRes{
		Code:          "200",
		Message:       "success",
		Configuration: cfg,
	}
	if data, err := cbor.Marshal(obj); err != nil {
		c.logError("marshal result failed", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		w.Header().Add("Content-Type", "application/cbor")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(data); err != nil {
			c.logError("write http body failed", err)
			return
		}
	}
}

func NewConfigureServer(opt ConfigureOptions) *ConfigureServer {
	versionComparator := opt.VersionComparator
	if versionComparator == nil {
		versionComparator = DefaultVersionComparator{}
	}

	// initialize server
	var srv = newServer(opt.DataPump, versionComparator)
	s := &ConfigureServer{
		opt:    opt,
		server: srv,
	}

	// read api
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	//r.Use(middleware.Logger) //FIXME require custom implementation
	// API - retrieve and listen
	r.Post("/retrieving", s.handleRetrieveAndListen)
	// API - get specific configuration
	r.Get("/configure/{group}/{key}", s.handleGetConfiguration)
	s.readMux = r

	// write api
	s.prepareWriteApi()

	return s
}

func (c *ConfigureServer) Startup() error {
	log.Println("ConfigureServer starting...")
	if err := c.server.Startup(); err != nil {
		return err
	}
	if err := c.startWriteApi(); err != nil {
		return err
	}
	// startup http server
	if httpServer, err := stdserver.StartCombinedStdHttpServer(&stdserver.CombinedStdHttpServerReq{
		Addr:    c.opt.Addr,
		TlsAddr: c.opt.TLSConfig.Addr,
		Cert:    c.opt.TLSConfig.Cert,
		CertKey: c.opt.TLSConfig.Key,
		Handler: c.readMux,
		StartedCallback: func(serverTypeName string) {
			log.Println("ConfigureServer [" + serverTypeName + "] endpoint started.")
		},
	}); err != nil {
		return err
	} else {
		c.httpServer = httpServer
	}
	return nil
}

func (c *ConfigureServer) Shutdown() error {
	if err := c.stopWriteApi(); err != nil {
		return err
	}
	// stop http server
	if c.httpServer != nil {
		if err := c.httpServer.Shutdown(context.Background()); err != nil {
			return err
		}
	}
	if c.server != nil {
		if err := c.server.Shutdown(); err != nil {
			return err
		}
	}
	return nil
}

func (c *ConfigureServer) prepareWriteApi() {
	if c.opt.WriteApi.DataWriter == nil {
		return
	}
	c.writeServer.writeServer = &writeServer{
		DataWriter: c.opt.WriteApi.DataWriter,
	}
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	//r.Use(middleware.Logger) //FIXME require custom implementation
	// save or update configuration
	r.Post("/configure", c.saveConfiguration)
	// delete configuration
	r.Delete("/configure/{group}/{key}", c.deleteConfiguration)

	c.writeServer.writeMux = r
}

func (c *ConfigureServer) startWriteApi() error {
	if c.writeServer.writeMux == nil {
		return nil
	}

	if err := c.writeServer.writeServer.Startup(); err != nil {
		return err
	}

	// startup http server
	if writeServer, err := stdserver.StartCombinedStdHttpServer(&stdserver.CombinedStdHttpServerReq{
		Addr:    c.opt.WriteApi.Addr,
		TlsAddr: c.opt.WriteApi.TLSConfig.Addr,
		Cert:    c.opt.WriteApi.TLSConfig.Cert,
		CertKey: c.opt.WriteApi.TLSConfig.Key,
		Handler: c.writeServer.writeMux,
		StartedCallback: func(serverTypeName string) {
			log.Println("ConfigureServer [" + serverTypeName + "] write endpoint started.")
		},
	}); err != nil {
		return err
	} else {
		c.writeServer.httpServer = writeServer
	}
	return nil
}

func (c *ConfigureServer) stopWriteApi() error {
	if c.writeServer.httpServer != nil {
		if err := c.writeServer.httpServer.Shutdown(context.Background()); err != nil {
			return err
		}
	}
	if c.writeServer.writeServer != nil {
		if err := c.writeServer.writeServer.Stop(); err != nil {
			return err
		}
	}
	return nil
}

func (c *ConfigureServer) saveConfiguration(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Accept") != "application/cbor" {
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}
	if r.Header.Get("Content-Type") != "application/cbor" {
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		c.logError("read http body error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	cfg := new(configapi.Configuration)
	if err := cbor.Unmarshal(data, cfg); err != nil {
		c.logError("parse http body error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if cfg.Group == "" || cfg.Key == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := c.writeServer.writeServer.SaveConfiguration(cfg); err != nil {
		c.logError("SaveConfiguration error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (c *ConfigureServer) deleteConfiguration(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Accept") != "application/cbor" {
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}
	selectorsInfo := strings.TrimSpace(r.Header.Get("X-Configuration-Sel"))
	if selectorsInfo == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	optSelectorsInfo := strings.TrimSpace(r.Header.Get("X-Configuration-Opt-Sel"))

	group := chi.URLParam(r, "group")
	if group == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	key := chi.URLParam(r, "key")
	if key == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ok, err := c.writeServer.writeServer.DeleteConfiguration(group, key, selectorsInfo, optSelectorsInfo)
	if err != nil {
		c.logError("DeleteConfiguration error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !ok {
		c.logWarn("no matching record deleted:", group, key)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
}
