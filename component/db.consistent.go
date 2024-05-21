package component

type _DbConsistentKv interface {
	Get(key string) (string, error)
	Set(key string, val string) error
	Del(key string) error
}

type _DbConsistentQuorum interface {
	// Leader responds the current node id of the leader
	Leader() (string, error)
	// Acquire tries to become the leader of the quorum
	// Responding the current node id of the leader, even of which is not current node
	// Note that this method may or may not block the process. It depends on the implementation.
	// Please try an infinitive loop to ensure the leader acquisition.
	// Acquisition success: parameter string == response string
	Acquire(node string) (string, error)
}

type ConsistentStore interface {
	_DbConsistentKv
	//_DbConsistentQuorum
}
