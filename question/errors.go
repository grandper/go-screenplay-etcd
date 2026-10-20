package question

import "errors"

var (
	// ErrKeyNotFound is wrapped by the error a question about one key returns
	// when the key does not exist in the store.
	ErrKeyNotFound = errors.New("was not found")

	// ErrKeyNotProvided is returned by a question about one key when the key is
	// empty.
	ErrKeyNotProvided = errors.New("the key must be provided")

	// ErrPrefixNotProvided is returned by a question about the keys starting
	// with a prefix when the prefix is empty.
	ErrPrefixNotProvided = errors.New("the prefix must be provided")

	// ErrNameNotProvided is returned by a question about an election when the
	// name is empty.
	ErrNameNotProvided = errors.New("the name must be provided")
)
