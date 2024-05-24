package comphttp

type ResponseHandler[T any] interface {
	Render(T) error
}

type HttpApi[REQ, RES any] interface {
	ParentUrl() string
	Url() string
	HttpMethod() []string
	Handle(r REQ) (ResponseHandler[RES], error)
}

type HttpApiSet[REQ, RES any] interface {
	AddHttpApi(a HttpApi[REQ, RES]) error

	DefaultErrorHandler(error) ResponseHandler[RES]
}
