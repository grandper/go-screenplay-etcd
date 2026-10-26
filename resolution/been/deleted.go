package been

import (
	"errors"

	"github.com/grandper/go-screenplay/v2/screenplay"

	"github.com/grandper/go-screenplay-etcd/etcd"
)

// Deleted creates a resolution to tell if an event removed its key, whatever
// the cause. Close it with ByTheExpiryOfItsLease to require the removal to be
// caused by the expiry of the lease of the key:
//
//	see.The(last.Of(events.OnTheKey("locks/batch-job"))).ShouldHave(been.Deleted())
//	see.The(last.Of(events.OnTheKey("locks/batch-job"))).ShouldHave(been.Deleted().ByTheExpiryOfItsLease())
//
// A nil event is not a delete.
func Deleted() *DeletedResolution {
	return &DeletedResolution{}
}

// DeletedResolution tells if an event removed its key.
type DeletedResolution struct {
	byTheExpiryOfItsLease bool
}

// ByTheExpiryOfItsLease requires the removal to be caused by the expiry (or
// the revocation) of the lease of the key, rather than by an actor.
func (r *DeletedResolution) ByTheExpiryOfItsLease() *DeletedResolution {
	r.byTheExpiryOfItsLease = true

	return r
}

// Inspect returns nil when the event removed its key, as required.
func (r *DeletedResolution) Inspect(event *etcd.Event) error {
	if event == nil {
		return errors.New("there is no event")
	}

	if !event.IsDelete() {
		return errors.New("it has been put")
	}

	if r.byTheExpiryOfItsLease && !event.CausedByLeaseExpiry() {
		return errors.New("it has been deleted by an actor, not by the expiry of its lease")
	}

	return nil
}

// String describes the resolution's expectation.
func (r *DeletedResolution) String() string {
	if r.byTheExpiryOfItsLease {
		return "been deleted by the expiry of its lease"
	}

	return "been deleted"
}

// DeletedResolution implements the screenplay.Resolution interface.
var _ screenplay.Resolution[*etcd.Event] = (*DeletedResolution)(nil)
