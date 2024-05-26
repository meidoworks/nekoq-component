package chi

import (
	"net/http"
	"testing"

	"github.com/meidoworks/nekoq-component/component/comphttp"
)

type Handler struct {
}

func (h Handler) ParentUrl() string {
	return "/aaaa"
}

func (h Handler) Url() string {
	return "/bbbb"
}

func (h Handler) HttpMethod() []string {
	//TODO implement me
	panic("implement me")
}

func (h Handler) Handle(r *http.Request) (comphttp.ResponseHandler[http.ResponseWriter], error) {
	//TODO implement me
	panic("implement me")
}

func TestMappingUrl(t *testing.T) {
	c := new(ChiHttpApiServer)
	u, err := c.mappingUrl(Handler{})
	if err != nil {
		t.Fatal(err)
	}
	if u != "/aaaa/bbbb" {
		t.Fatal("unexpected url")
	}
}
