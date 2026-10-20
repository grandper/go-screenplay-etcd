package ability

import (
	"errors"

	"go.etcd.io/etcd/client/v3/concurrency"
)

var (
	// ErrNoEndpointProvided is returned by the first operation of an ability
	// that was given neither an endpoint nor a client.
	ErrNoEndpointProvided = errors.New("no etcd endpoint was provided to the ability")

	// ErrFailedToConnect is wrapped by the error returned when the client of
	// the etcd server cannot be created, or when a watch is canceled by the
	// server.
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
