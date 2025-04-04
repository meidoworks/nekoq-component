package chi2

import (
	"errors"
	"log"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-chi/chi/v5"
)

var (
	ErrFieldNotFound     = errors.New("field not found")
	ErrFieldTypeMismatch = errors.New("field type mismatch")
	ErrMethodInvalid     = errors.New("method is invalid")
	ErrEmptyURL          = errors.New("url is empty")
)

type ChiApiStub struct {
	items map[string]map[string]struct {
		m string
		u string
		Middlewares
		RequestValidators
		internalCheck
	}
}

func NewChiApiStub() *ChiApiStub {
	return &ChiApiStub{
		items: map[string]map[string]struct {
			m string
			u string
			Middlewares
			RequestValidators
			internalCheck
		}{},
	}
}

func (c *ChiApiStub) addItem(method, url string, mw Middlewares, rv RequestValidators, ct internalCheck) {
	sub, ok := c.items[method]
	if !ok {
		sub = map[string]struct {
			m string
			u string
			Middlewares
			RequestValidators
			internalCheck
		}{}
		c.items[method] = sub
	}
	item, ok := sub[url]
	if !ok {
		item = struct {
			m string
			u string
			Middlewares
			RequestValidators
			internalCheck
		}{m: method, u: url, Middlewares: mw, RequestValidators: rv, internalCheck: ct}
		sub[url] = item
	} else {
		item.Middlewares = mw
		item.RequestValidators = rv
		item.internalCheck = ct
	}
}

func (c *ChiApiStub) RegisterControllers(controllers ...internalCheck) error {
	for _, controller := range controllers {
		var m string
		var u string
		if method, err := extractMethod(controller); err != nil {
			return err
		} else {
			m = method
		}
		if url, err := extractUrl(controller); err != nil {
			return err
		} else {
			u = url
		}
		c.addItem(m, u, controller.middlewares(), controller.requestValidators(), controller)
		return nil
	}
	return nil
}

func extractUrl(controller internalCheck) (string, error) {
	ct := reflect.TypeOf(controller)
	isPtr := false
	if ct.Kind() == reflect.Ptr {
		isPtr = true
		ct = ct.Elem()
	}
	var _ = isPtr

	field, ok := ct.FieldByName("URL")
	if !ok {
		return "", ErrFieldNotFound
	}
	if field.Type.Name() != "string" {
		return "", ErrFieldTypeMismatch
	}
	method := field.Tag.Get("url")
	method = strings.TrimSpace(method)
	if method == "" {
		return "", ErrEmptyURL
	}
	return method, nil
}

func extractMethod(controller internalCheck) (string, error) {
	ct := reflect.TypeOf(controller)
	isPtr := false
	if ct.Kind() == reflect.Ptr {
		isPtr = true
		ct = ct.Elem()
	}
	var _ = isPtr

	field, ok := ct.FieldByName("Method")
	if !ok {
		return "", ErrFieldNotFound
	}
	if field.Type.Name() != "string" {
		return "", ErrFieldTypeMismatch
	}
	method := field.Tag.Get("method")
	method = strings.ToUpper(strings.TrimSpace(method))
	if _, ok := httpMethods[method]; !ok {
		return "", ErrMethodInvalid
	}
	return method, nil
}

func (c *ChiApiStub) LogAllControllers() {
	for method, sub := range c.items {
		log.Println("====>>>> method:", method)
		for url, item := range sub {
			log.Println(url, len(item.Middlewares), len(item.RequestValidators))
		}
	}
}

func (c *ChiApiStub) BuildFor(r *chi.Mux) {
	for method, sub := range c.items {
		for url, item := range sub {
			newItem := item
			r.With(item.Middlewares...).MethodFunc(method, url, c.generalHandler(newItem))
		}
	}
}

func (c *ChiApiStub) generalHandler(item struct {
	m string
	u string
	Middlewares
	RequestValidators
	internalCheck
}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		for _, rv := range item.RequestValidators {
			render := rv(w, r)
			if render != nil {
				if err := render.Render(w, r); err != nil {
					//FIXME handling render error
				}
				return // break
			}
		}

		render := item.internalCheck.HandleHttp(w, r)
		//FIXME perhaps do render nil check
		if err := render.Render(w, r); err != nil {
			//FIXME handling render error
		}
	}
}
