package release

import (
	"github.com/grandper/go-screenplay/v2/screenplay"

	"github.com/grandper/go-screenplay-etcd/etcd"
	impl "github.com/grandper/go-screenplay-etcd/internal/ability"
	"github.com/grandper/go-screenplay-etcd/internal/verdict"
)

// TheLock releases a distributed lock the actor holds:
//
//	release.TheLock("deployment")
func TheLock(name string) screenplay.Performable {
	return &theLockAction{name: name}
}

// theLockAction releases a distributed lock.
type theLockAction struct {
	name string
}

// String describes the action.
func (a *theLockAction) String() string {
	return "release the lock " + verdict.Describe(a.name)
}

// PerformAs releases the lock.
func (a *theLockAction) PerformAs(theActor *screenplay.Actor) error {
	if a.name == "" {
		return etcd.ErrNameNotProvided
	}

	return impl.TheAbilityOf(theActor).ToReleaseTheLock(theActor.Context(), a.name)
}

// theLockAction implements the screenplay.Performable interface.
var _ screenplay.Performable = (*theLockAction)(nil)
