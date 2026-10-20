package ability

import "errors"

var (
	// ErrNoEndpointProvided is returned by the first operation of an ability
	// that was given neither an endpoint nor a client.
	ErrNoEndpointProvided = errors.New("no etcd endpoint was provided to the ability")

	// ErrFailedToConnect is wrapped by the error returned when the client of
	// the etcd server cannot be created.
	ErrFailedToConnect = errors.New("failed to connect to the etcd server")
)
