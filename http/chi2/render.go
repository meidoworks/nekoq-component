package chi2

import (
	"encoding/json"
	"net/http"
)

type statusRender struct {
	code int
}

func (s statusRender) Render(w http.ResponseWriter, r *http.Request) error {
	w.WriteHeader(s.code)
	return nil
}

func NewStatusRender(code int) Render {
	return statusRender{code: code}
}

type jsonObjRender struct {
	code int
	obj  any
}

func (j jsonObjRender) Render(w http.ResponseWriter, r *http.Request) error {
	data, err := json.Marshal(j.obj)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(j.code)
	_, err = w.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func NewJsonOkRender(obj any) Render {
	return jsonObjRender{code: http.StatusOK, obj: obj}
}

func NewJsonObjRender(code int, obj any) Render {
	return jsonObjRender{code: code, obj: obj}
}

type errRender struct {
	err error
}

func (e errRender) Render(w http.ResponseWriter, r *http.Request) error {
	w.WriteHeader(http.StatusInternalServerError)
	//FIXME adding custom header and body error type
	return e.err
}

func NewErrRender(err error) Render {
	return errRender{err: err}
}
