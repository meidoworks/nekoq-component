package chi

import (
	"log"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"

	"github.com/meidoworks/nekoq-component/component/comphttp"
)

type defaultErrorRender struct {
	err error
}

func (d defaultErrorRender) Render(t http.ResponseWriter) error {
	t.WriteHeader(http.StatusInternalServerError)
	return nil
}

type ChiHttpApiServerConfig struct {
	Addr string
}

type ChiHttpApiServer struct {
	handlerList []comphttp.HttpApi[*http.Request, http.ResponseWriter]

	chiRouter *chi.Mux
	cfg       *ChiHttpApiServerConfig
}

func (c *ChiHttpApiServer) StartServing() error {
	return http.ListenAndServe(c.cfg.Addr, c.chiRouter)
}

func (c *ChiHttpApiServer) DefaultErrorHandler(err error) comphttp.ResponseHandler[http.ResponseWriter] {
	return defaultErrorRender{err}
}

func (c *ChiHttpApiServer) AddHttpApi(a comphttp.HttpApi[*http.Request, http.ResponseWriter]) error {
	u, err := url.Parse(a.ParentUrl() + "/" + a.Url())
	if err != nil {
		return err
	}
	base, err := url.Parse("/")
	if err != nil {
		log.Fatal(err)
	}

	fullPath := base.ResolveReference(u).Path

	for _, method := range a.HttpMethod() {
		c.chiRouter.MethodFunc(method, fullPath, func(writer http.ResponseWriter, request *http.Request) {
			var renderErr error
			render, err := a.Handle(request)
			if err != nil {
				renderErr = c.DefaultErrorHandler(err).Render(writer)
			} else {
				renderErr = render.Render(writer)
			}
			if renderErr != nil {
				//FIXME try to find a way to handler render error.
				//Note that: the writer may has been committed with an http status.
			}
		})
	}

	c.handlerList = append(c.handlerList, a)
	return nil
}

var _ comphttp.HttpApiSet[*http.Request, http.ResponseWriter] = new(ChiHttpApiServer)

func NewChiHttpApiServer(cfg *ChiHttpApiServerConfig) *ChiHttpApiServer {
	r := chi.NewRouter()
	return &ChiHttpApiServer{
		chiRouter: r,
		cfg:       cfg,
	}
}
