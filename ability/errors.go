package ability

import impl "github.com/grandper/go-screenplay-etcd/internal/ability"

var (
	// ErrNoEndpointProvided is returned by the first action of an actor whose
	// ability was given neither an endpoint nor a client.
	ErrNoEndpointProvided = impl.ErrNoEndpointProvided

	// ErrFailedToConnect is wrapped by the error returned when the client of
	// the etcd server cannot be created, or when a watch is canceled by the
	// server.
	ErrFailedToConnect = impl.ErrFailedToConnect

	// ErrAlreadyWatching is wrapped by the error returned when the actor
	// watches a key, or the keys starting with a prefix, it already watches.
	ErrAlreadyWatching = impl.ErrAlreadyWatching

	// ErrNotWatching is wrapped by the error returned when the actor asks about
	// the events on a key, or on the keys starting with a prefix, it does not
	// watch.
	ErrNotWatching = impl.ErrNotWatching
)
