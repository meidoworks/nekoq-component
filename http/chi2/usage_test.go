package chi2

import (
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type DemoController struct {
	Controller
	Method string `method:"GET"`
	URL    string `url:"/api/v1/user/{userId}/detail"`
}

func (d *DemoController) HandleHttp(w http.ResponseWriter, r *http.Request) Render {
	//TODO implement me
	panic("implement me")
}

func NewDemoController() *DemoController {
	return &DemoController{
		Controller: Controller{
			Middlewares: Middlewares{
				middleware.Logger,
				middleware.RealIP,
				middleware.Recoverer,
			},
			RequestValidators: RequestValidators{
				AcceptContentTypeFor("application/json"),
			},
		},
	}
}

func TestBasicUsage(t *testing.T) {
	r := chi.NewMux()

	chi2 := NewChiApiStub()
	if err := chi2.RegisterControllers(NewDemoController()); err != nil {
		t.Fatal(err)
	}

	chi2.LogAllControllers()

	chi2.BuildFor(r)

	//http.ListenAndServe(":3333", r)
	t.Log("done")
}
