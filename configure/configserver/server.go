package configserver

import (
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/meidoworks/nekoq-component/configure/configapi"
)

type ConfigureOptions struct {
	Addr      string
	TLSConfig struct {
		Addr string
		Cert string
		Key  string
	}

	MaxWaitTimeForUpdate int // in seconds

	DataPump configapi.DataPump
}

func (c *ConfigureOptions) GetMaxWaitTimeForUpdate() time.Duration {
	if c.MaxWaitTimeForUpdate <= 0 {
		return 60 * time.Second
	} else {
		return time.Duration(c.MaxWaitTimeForUpdate) * time.Second
	}
}

type ConfigureServer struct {
	readMux  *chi.Mux // for client read
	writeMux *chi.Mux // for management portal write

	server *server
	opt    ConfigureOptions

	httpServer *http.Server
}

func (c *ConfigureServer) logError(messsage string, err error) {
	if err != nil {
		log.Println("[ERROR]", messsage, err)
	} else {
		log.Println("[ERROR]", messsage)
	}
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
	// initialize server
	var srv = newServer(opt.DataPump)
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

	//TODO write api

	return s
}

func (c *ConfigureServer) Startup() error {
	if err := c.server.Startup(); err != nil {
		return err
	}
	l, err := net.Listen("tcp", c.opt.Addr)
	if err != nil {
		return err
	}
	// startup http server
	srv := &http.Server{Handler: c.readMux}
	c.httpServer = srv
	go func() {
		if err := srv.Serve(l); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				panic(err)
			}
		}
	}()
	return nil
}

func (c *ConfigureServer) Shutdown() error {
	// stop http server
	if err := c.httpServer.Close(); err != nil {
		return err
	}
	if err := c.server.Shutdown(); err != nil {
		return err
	}
	return nil
}
