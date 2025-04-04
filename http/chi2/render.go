package chi2

import (
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
