package chi

import (
	"encoding/json"
	"net/http"

	"github.com/meidoworks/nekoq-component/component/comphttp"
)

type chiRenderString struct {
	status int
	raw    string
}

func (c *chiRenderString) Render(w http.ResponseWriter) error {
	w.WriteHeader(c.status)
	if _, err := w.Write([]byte(c.raw)); err != nil {
		return err
	} else {
		return nil
	}
}

func RenderString(status int, raw string) comphttp.ResponseHandler[http.ResponseWriter] {
	return &chiRenderString{raw: raw, status: status}
}

type chiRenderStatus struct {
	status int
}

func (c chiRenderStatus) Render(w http.ResponseWriter) error {
	w.WriteHeader(c.status)
	return nil
}

func RenderStatus(status int) comphttp.ResponseHandler[http.ResponseWriter] {
	return chiRenderStatus{status: status}
}

type chiRenderOKString struct {
	str string
}

func (c *chiRenderOKString) Render(w http.ResponseWriter) error {
	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "text/plain")
	_, err := w.Write([]byte(c.str))
	return err
}

func RenderOKString(str string) comphttp.ResponseHandler[http.ResponseWriter] {
	return &chiRenderOKString{str: str}
}

type chiRenderError struct {
	err error
}

func (c *chiRenderError) Render(w http.ResponseWriter) error {
	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Add("Content-Type", "text/plain")
	_, err := w.Write([]byte(c.err.Error()))
	return err
}

func RenderError(err error) comphttp.ResponseHandler[http.ResponseWriter] {
	return &chiRenderError{err: err}
}

type chiRenderJson struct {
	status int
	obj    interface{}
}

func (c *chiRenderJson) Render(w http.ResponseWriter) error {
	data, err := json.Marshal(c.obj)
	if err != nil {
		return err
	}
	w.WriteHeader(c.status)
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	_, err = w.Write(data)
	return err
}

func RenderJson(status int, obj interface{}) comphttp.ResponseHandler[http.ResponseWriter] {
	return &chiRenderJson{obj: obj, status: status}
}

type chiRenderBinary struct {
	status int
	data   []byte
}

func (c *chiRenderBinary) Render(w http.ResponseWriter) error {
	w.WriteHeader(c.status)
	w.Header().Add("Content-Type", "application/octet-stream")
	_, err := w.Write(c.data)
	return err
}

func RenderBinary(status int, data []byte) comphttp.ResponseHandler[http.ResponseWriter] {
	return &chiRenderBinary{status: status, data: data}
}
