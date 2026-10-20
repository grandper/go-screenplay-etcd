package ability

import (
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// Event is a change the actor observed on a watched key: a put, or a delete.
// Its parts are read through its accessors.
type Event struct {
	key                 string
	value               string
	previousValue       string
	deleted             bool
	revision            int64
	version             int64
	causedByLeaseExpiry bool
}

// NewEvent creates the event the store notified, telling whether a delete was
// caused by the expiry of the lease of the key.
func NewEvent(event *clientv3.Event, causedByLeaseExpiry bool) *Event {
	return &Event{
		key:                 string(event.GetKv().GetKey()),
		value:               string(event.GetKv().GetValue()),
		previousValue:       string(event.GetPrevKv().GetValue()),
		deleted:             event.GetType() == mvccpb.DELETE,
		revision:            event.GetKv().GetModRevision(),
		version:             event.GetKv().GetVersion(),
		causedByLeaseExpiry: causedByLeaseExpiry,
	}
}

// Key returns the key the event happened to.
func (e *Event) Key() string {
	return e.key
}

// IsPut returns true when the event put a value for the key.
func (e *Event) IsPut() bool {
	return !e.deleted
}

// IsDelete returns true when the event removed the key.
func (e *Event) IsDelete() bool {
	return e.deleted
}

// Value returns the value the event put for the key, empty for a delete.
func (e *Event) Value() string {
	return e.value
}

// PreviousValue returns the value the key held before the event, empty when
// the event created the key.
func (e *Event) PreviousValue() string {
	return e.previousValue
}

// Revision returns the revision of the store at which the event happened.
func (e *Event) Revision() int64 {
	return e.revision
}

// Version returns the version of the key after the event: the number of
// times it was put since it was created, zero after a delete.
func (e *Event) Version() int64 {
	return e.version
}

// CausedByLeaseExpiry returns true when the event removed the key because its
// lease expired or was revoked, rather than because an actor removed it.
func (e *Event) CausedByLeaseExpiry() bool {
	return e.causedByLeaseExpiry
}
