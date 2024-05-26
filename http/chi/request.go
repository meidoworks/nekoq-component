package chi

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func GetUrlParam(r *http.Request, key string) string {
	return chi.URLParam(r, key)
}

func BindJson(r *http.Request, obj any) error {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, obj)
}
