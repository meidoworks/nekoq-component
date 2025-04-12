package chi2

import (
	"errors"
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
	v1 any // v1 mechanisms
	Middlewares
	RequestValidators
	BodyParser func(req *http.Request, r io.Reader) (any, error)

	v2                   any // +v2 mechanisms
	BodyVerifier         func(req *http.Request, r any) Render
	NewCustomErrorRender func(error) Render
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
	return c.BodyParser(r, r.Body)
}

func (c Controller) HandleHttp(w http.ResponseWriter, r *http.Request) Render {
	var model any
	if obj, err := c.ParseBody(r); err != nil {
		if c.NewCustomErrorRender == nil {
			return NewErrRender(err)
		} else {
			return c.NewCustomErrorRender(err)
		}
	} else {
		model = obj
	}
	if c.BodyVerifier == nil {
		err := errors.New("body verifier is nil")
		if c.NewCustomErrorRender == nil {
			return NewErrRender(err)
		} else {
			return c.NewCustomErrorRender(err)
		}
	} else if r := c.BodyVerifier(r, model); r != nil {
		return r
	}
	return c.HandleModel(model, w, r)
}

func (c Controller) HandleModel(model any, w http.ResponseWriter, r *http.Request) Render {
	panic("implement me")
}

type Middlewares []func(http.Handler) http.Handler

type RequestValidators []func(w http.ResponseWriter, r *http.Request) Render

type Render interface {
	Render(w http.ResponseWriter, r *http.Request) error
}
