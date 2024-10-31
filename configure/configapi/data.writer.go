package configapi

type DataWriter interface {
	Startup() error
	Stop() error
	SaveConfiguration(cfg Configuration) error
}
