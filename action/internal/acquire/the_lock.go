package acquire

import (
	"github.com/grandper/go-screenplay/v2/screenplay"

	"github.com/grandper/go-screenplay-etcd/etcd"
	impl "github.com/grandper/go-screenplay-etcd/internal/ability"
	"github.com/grandper/go-screenplay-etcd/internal/verdict"
)

// TheLock acquires a distributed lock, waiting for the actor holding it to
// release it. The lock is held until the actor releases it with
// release.TheLock, or exits:
//
//	acquire.TheLock("deployment")
func TheLock(name string) *TheLockAction {
	return &TheLockAction{name: name}
}

// TheLockAction acquires a distributed lock.
type TheLockAction struct {
	name           string
	withoutWaiting bool
}

// WithoutWaiting makes the action fail, with etcd.ErrLocked, when another
// actor holds the lock, instead of waiting for it; it reads "without waiting".
func (a *TheLockAction) WithoutWaiting() *TheLockAction {
	a.withoutWaiting = true

	return a
}

// String describes the action.
func (a *TheLockAction) String() string {
	description := "acquire the lock " + verdict.Describe(a.name)
	if a.withoutWaiting {
		description += " without waiting"
	}

	return description
}

// PerformAs acquires the lock.
func (a *TheLockAction) PerformAs(theActor *screenplay.Actor) error {
	if a.name == "" {
		return etcd.ErrNameNotProvided
	}

	return impl.TheAbilityOf(theActor).ToAcquireTheLock(theActor.Context(), a.name, a.withoutWaiting)
}

// TheLockAction implements the screenplay.Performable interface.
var _ screenplay.Performable = (*TheLockAction)(nil)
