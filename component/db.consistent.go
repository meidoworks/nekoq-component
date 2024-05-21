package component

type _DbConsistentKv interface {
	Get(key string) (string, error)
	Set(key string, val string) error
	Del(key string) error
}

type _DbConsistentQuorum interface {
	// Leader responds the current node id of the leader
	Leader(key string) (string, error)
	// Acquire tries to become the leader of the quorum
	// Responding the current node id of the leader, even of which is not current node
	// Note that this method may or may not block the process. It depends on the implementation.
	// Please try an infinitive loop to ensure the leader acquisition.
	// Acquisition success: parameter string == response string
	Acquire(key, node string) (string, error)
}

type _DbConsistentWatch interface {
	WatchFolder(folder string) (<-chan WatchEvent, CancelFn, error)
}

type ConsistentStore interface {
	_DbConsistentKv
	_DbConsistentQuorum
	_DbConsistentWatch
}

type WatchEvent struct {
	Path string
	Ev   []struct {
		Key       string
		EventType WatchEventType
	}
}

type WatchEventType int

const (
	WatchEventUnknown WatchEventType = iota
	WatchEventCreated
	WatchEventModified
	WatchEventDelete
)
