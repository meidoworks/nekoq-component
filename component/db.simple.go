package component

import "errors"

var (
	ErrUnsupportedOperation = errors.New("unsupported operation")
	ErrDuplicatedObjectById = errors.New("duplicated object by id")
)

type SimpleStore interface {
	Table(table string) SimpleStoreTable
}

type SimpleStoreTable interface {
	// QueryById accepts an empty object to be filled with data by the given id
	// If no object is found, then the nil object will be returned.
	QueryById(id []byte, empty SimpleStoreObject) (SimpleStoreObject, error)
	Insert(obj SimpleStoreObject) error
	Delete(id []byte) error
	Update(obj SimpleStoreObject) error
}

type SimpleStoreObject interface {
	Id() []byte
	//IndexElements() []SimpleStoreObjectIndexElement

	Marshal() ([]byte, error)
	Unmarshal(data []byte) error
}

type SimpleStoreObjectIndexElement interface {
	IndexName() string
	Value() []byte
	Unique() bool
}
