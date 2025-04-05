package chi2

import (
	"net/http"
	"strings"
)

func AllowContentTypeFor(contentTypes ...string) func(w http.ResponseWriter, r *http.Request) Render {
	ctm := map[string]struct{}{}
	for _, ct := range contentTypes {
		ctm[ct] = struct{}{}
	}
	return func(w http.ResponseWriter, r *http.Request) Render {
		ct := r.Header.Get("Content-Type")
		ct = strings.TrimSpace(strings.Split(ct, ";")[0])
		_, ok := ctm[ct]
		if ok {
			return nil
		}
		return NewStatusRender(http.StatusUnsupportedMediaType)
	}
}
