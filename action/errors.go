package action

import "errors"

var (
	// ErrKeyNotProvided is returned when an action is performed without a key,
	// or with an empty one. Nothing panics when the action is built; the mistake
	// is reported when the actor performs it.
	ErrKeyNotProvided = errors.New("the key must be provided")

	// ErrPrefixNotProvided is returned when an action about the keys starting
	// with a prefix is performed with an empty prefix, which would reach every
	// key of the store.
	ErrPrefixNotProvided = errors.New("the prefix must be provided")
)
