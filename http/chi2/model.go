package chi2

import (
	"io"
	"net/http"
)

var (
	httpMethods = map[string]struct{}{
		http.MethodGet:     {},
		http.MethodPost:    {},
		http.MethodPut:     {},
		http.MethodConnect: {},
		http.MethodDelete:  {},
		http.MethodHead:    {},
		http.MethodOptions: {},
		http.MethodPatch:   {},
		http.MethodTrace:   {},
	}
)

type HttpController interface {
	HandleHttp(w http.ResponseWriter, r *http.Request) Render
}

type internalCheck interface {
	_chi_internal1_779960()
	_chi_internal2_988103()
	_chi_internal3_295800()

	middlewares() Middlewares
	requestValidators() RequestValidators

	HttpController
}

type Controller struct {
	Middlewares
	RequestValidators
	BodyParser func(r io.Reader) (any, error)
}

func (c Controller) middlewares() Middlewares {
	return c.Middlewares
}

func (c Controller) requestValidators() RequestValidators {
	return c.RequestValidators
}

func (c Controller) _chi_internal1_779960() {
	panic("implement me")
}

func (c Controller) _chi_internal2_988103() {
	panic("implement me")
}

func (c Controller) _chi_internal3_295800() {
	panic("implement me")
}

func (c Controller) ParseBody(r *http.Request) (any, error) {
	return c.BodyParser(r.Body)
}

type Middlewares []func(http.Handler) http.Handler

type RequestValidators []func(w http.ResponseWriter, r *http.Request) Render

type Render interface {
	Render(w http.ResponseWriter, r *http.Request) error
}
