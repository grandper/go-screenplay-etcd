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

	// ErrKeyAlreadyExists is wrapped by the error put.TheValue returns when the
	// value was to be put only if the key was absent, and the key exists.
	ErrKeyAlreadyExists = errors.New("already exists")

	// ErrKeyModified is wrapped by the error put.TheValue and remove.TheKey
	// return when the key was to be written only if it was unchanged since a
	// revision, and it was modified.
	ErrKeyModified = errors.New("was modified")
)
