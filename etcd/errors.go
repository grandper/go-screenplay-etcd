package etcd

import (
	"errors"

	"go.etcd.io/etcd/client/v3/concurrency"
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

	// ErrTimeToLiveNotProvided is returned when put.TheValue is to keep the
	// key alive, with AndKeepItAlive, but was given no time to live with
	// ExpiringAfter: a key that never expires has no lease to renew.
	ErrTimeToLiveNotProvided = errors.New("the time to live must be provided to keep the key alive")
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
	ErrNoEndpointProvided = errors.New("no etcd endpoint was provided to the ability")

	// ErrFailedToConnect is wrapped by the error returned when the client of
	// the etcd server cannot be created, when the connection is not ready
	// within the timeout of the ability (five seconds without one), or when
	// a watch is canceled by the server.
	ErrFailedToConnect = errors.New("failed to connect to the etcd server")

	// ErrFailedToLoadTheCertificates is wrapped by the error returned when the
	// certificate authority, or the client certificate and its key, given to
	// the ability cannot be loaded.
	ErrFailedToLoadTheCertificates = errors.New("failed to load the certificates")

	// ErrAlreadyWatching is wrapped by the error returned when the actor
	// watches a key, or the keys starting with a prefix, it already watches.
	ErrAlreadyWatching = errors.New("already watched")

	// ErrNotWatching is wrapped by the error returned when the actor asks about
	// the events on a key, or on the keys starting with a prefix, it does not
	// watch.
	ErrNotWatching = errors.New("not watched")

	// ErrLocked is returned when the actor acquires, without waiting, a lock
	// another actor holds. It is the error of the etcd client.
	ErrLocked = concurrency.ErrLocked

	// ErrLockAlreadyHeld is wrapped by the error returned when the actor
	// acquires a lock it already holds.
	ErrLockAlreadyHeld = errors.New("already held by the actor")

	// ErrLockNotHeld is wrapped by the error returned when the actor releases a
	// lock it does not hold.
	ErrLockNotHeld = errors.New("not held by the actor")

	// ErrAlreadyCampaigning is wrapped by the error returned when the actor
	// campaigns in an election it already campaigns in.
	ErrAlreadyCampaigning = errors.New("the actor already campaigns")

	// ErrNotCampaigning is wrapped by the error returned when the actor resigns
	// from an election it does not campaign in.
	ErrNotCampaigning = errors.New("the actor does not campaign")
)
