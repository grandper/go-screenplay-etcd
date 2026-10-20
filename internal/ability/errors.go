package ability

import "errors"

var (
	// ErrNoEndpointProvided is returned by the first operation of an ability
	// that was given neither an endpoint nor a client.
	ErrNoEndpointProvided = errors.New("no etcd endpoint was provided to the ability")

	// ErrFailedToConnect is wrapped by the error returned when the client of
	// the etcd server cannot be created, or when a watch is canceled by the
	// server.
	ErrFailedToConnect = errors.New("failed to connect to the etcd server")

	// ErrAlreadyWatching is wrapped by the error returned when the actor
	// watches a key, or the keys starting with a prefix, it already watches.
	ErrAlreadyWatching = errors.New("already watched")

	// ErrNotWatching is wrapped by the error returned when the actor asks about
	// the events on a key, or on the keys starting with a prefix, it does not
	// watch.
	ErrNotWatching = errors.New("not watched")
)
