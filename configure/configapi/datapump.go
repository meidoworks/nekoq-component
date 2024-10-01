package configapi

type Event struct {
	Configuration *Configuration

	Created  bool
	Modified bool
	Deleted  bool
}

type DataPump interface {
	Startup() error
	Stop() error
	EventChannel() <-chan Event
	TriggerDumpToChannel() <-chan Event
}
