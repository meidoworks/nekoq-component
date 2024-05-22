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
	// WatchFolder support watching a prefix
	// After trigger this command, a full fresh of the prefix will be retrieved at once.
	// Then any changes happen inside the prefix will be notified, including children change, children value change.
	// List of possible changes:
	// * Child creation
	// * Child deletion
	// * Child value change
	WatchFolder(folder string) (<-chan WatchEvent, CancelFn, error)
}

type ConsistentStore interface {
	_DbConsistentKv
	_DbConsistentQuorum
	_DbConsistentWatch
}

type WatchEvent struct {
	// Path is the full path of the element
	Path string
	Ev   []struct {
		Key       string
		EventType WatchEventType
	}
}

type WatchEventType int

const (
	WatchEventUnknown WatchEventType = iota
	WatchEventFresh
	WatchEventCreated
	WatchEventModified
	WatchEventDelete
)
