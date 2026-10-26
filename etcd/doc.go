// Package etcd holds what the actor who can use etcd, given with use.ETCD,
// answers with and reports: a Key, an Event, an EndpointStatus, an Alarm, and
// the errors of the ability:
//
//	theKey, err := anActor.AsksFor(key.Named("app/status")) // theKey is an *etcd.Key
//	errors.Is(err, etcd.ErrNotWatching)
package etcd
