package etcd

import (
	"errors"

	impl "github.com/grandper/go-screenplay-etcd/internal/ability"
)

// The errors reported when a statement is mistaken. Nothing panics when an
// action or a question is built; the mistake is reported when the actor
// performs the action or asks the question.
var (
	// ErrKeyNotProvided is returned when an action or a question about one key
	// is performed without a key, or with an empty one.
	ErrKeyNotProvided = errors.New("the key must be provided")

	// ErrPrefixNotProvided is returned when an action or a question about the
	// keys starting with a prefix is performed with an empty prefix, which
	// would reach every key of the store.
	ErrPrefixNotProvided = errors.New("the prefix must be provided")

	// ErrNameNotProvided is returned when an action or a question about a lock
	// or an election is performed with an empty name.
	ErrNameNotProvided = errors.New("the name must be provided")

	// ErrEndpointNotProvided is returned when a question about an endpoint is
	// asked with an empty endpoint.
	ErrEndpointNotProvided = errors.New("the endpoint must be provided")
)

// The errors reported by the actions and the questions about the keys.
var (
	// ErrKeyNotFound is wrapped by the error a question about one key returns
	// when the key does not exist in the store.
	ErrKeyNotFound = errors.New("was not found")

	// ErrKeyAlreadyExists is wrapped by the error put.TheValue returns when the
	// value was to be put only if the key was absent, and the key exists.
	ErrKeyAlreadyExists = errors.New("already exists")

	// ErrKeyModified is wrapped by the error put.TheValue and remove.TheKey
	// return when the key was to be written only if it was unchanged since a
	// revision, and it was modified.
	ErrKeyModified = errors.New("was modified")
)

// The errors reported by the ability.
var (
	// ErrNoEndpointProvided is returned by the first action of an actor whose
	// ability was given neither an endpoint nor a client.
	ErrNoEndpointProvided = impl.ErrNoEndpointProvided

	// ErrFailedToConnect is wrapped by the error returned when the client of
	// the etcd server cannot be created, or when a watch is canceled by the
	// server.
	ErrFailedToConnect = impl.ErrFailedToConnect

	// ErrFailedToLoadTheCertificates is wrapped by the error returned when the
	// certificate authority, or the client certificate and its key, given to
	// the ability cannot be loaded.
	ErrFailedToLoadTheCertificates = impl.ErrFailedToLoadTheCertificates

	// ErrAlreadyWatching is wrapped by the error returned when the actor
	// watches a key, or the keys starting with a prefix, it already watches.
	ErrAlreadyWatching = impl.ErrAlreadyWatching

	// ErrNotWatching is wrapped by the error returned when the actor asks about
	// the events on a key, or on the keys starting with a prefix, it does not
	// watch.
	ErrNotWatching = impl.ErrNotWatching

	// ErrLocked is returned when the actor acquires, without waiting, a lock
	// another actor holds. It is the error of the etcd client.
	ErrLocked = impl.ErrLocked

	// ErrLockAlreadyHeld is wrapped by the error returned when the actor
	// acquires a lock it already holds.
	ErrLockAlreadyHeld = impl.ErrLockAlreadyHeld

	// ErrLockNotHeld is wrapped by the error returned when the actor releases a
	// lock it does not hold.
	ErrLockNotHeld = impl.ErrLockNotHeld

	// ErrAlreadyCampaigning is wrapped by the error returned when the actor
	// campaigns in an election it already campaigns in.
	ErrAlreadyCampaigning = impl.ErrAlreadyCampaigning

	// ErrNotCampaigning is wrapped by the error returned when the actor resigns
	// from an election it does not campaign in.
	ErrNotCampaigning = impl.ErrNotCampaigning
)
