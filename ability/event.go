package ability

import impl "github.com/grandper/go-screenplay-etcd/internal/ability"

// Event is a change the actor observed on a watched key: a put, or a delete.
// Its parts are read through Key(), IsPut(), IsDelete(), Value(),
// PreviousValue(), Revision(), Version() and CausedByLeaseExpiry().
type Event = impl.Event
