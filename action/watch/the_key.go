package watch

import (
	"github.com/grandper/go-screenplay/v2/screenplay"

	"github.com/grandper/go-screenplay-etcd/action"
	impl "github.com/grandper/go-screenplay-etcd/internal/ability"
	"github.com/grandper/go-screenplay-etcd/internal/verdict"
)

// TheKey watches a key, recording every change made to it until the actor
// exits. The action returns once the store confirmed the watch, so that a
// change made right after it is recorded:
//
//	watch.TheKey("app/config/database")
func TheKey(key string) screenplay.Performable {
	return &theKeyAction{key: key}
}

// theKeyAction watches a key.
type theKeyAction struct {
	key string
}

// String describes the action.
func (a *theKeyAction) String() string {
	return "watch the key " + verdict.Describe(a.key)
}

// PerformAs watches the key.
func (a *theKeyAction) PerformAs(theActor *screenplay.Actor) error {
	if a.key == "" {
		return action.ErrKeyNotProvided
	}

	return impl.TheAbilityOf(theActor).ToWatch(theActor.Context(), a.key, false)
}

// theKeyAction implements the screenplay.Performable interface.
var _ screenplay.Performable = (*theKeyAction)(nil)
