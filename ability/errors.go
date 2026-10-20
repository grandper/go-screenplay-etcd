package ability

import impl "github.com/grandper/go-screenplay-etcd/internal/ability"

var (
	// ErrNoEndpointProvided is returned by the first action of an actor whose
	// ability was given neither an endpoint nor a client.
	ErrNoEndpointProvided = impl.ErrNoEndpointProvided

	// ErrFailedToConnect is wrapped by the error returned when the client of
	// the etcd server cannot be created.
	ErrFailedToConnect = impl.ErrFailedToConnect
)
