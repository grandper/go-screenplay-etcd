package remove

import (
	"github.com/grandper/go-screenplay/v2/screenplay"

	"github.com/grandper/go-screenplay-etcd/action"
	impl "github.com/grandper/go-screenplay-etcd/internal/ability"
	"github.com/grandper/go-screenplay-etcd/internal/verdict"
)

// TheKey removes a key from the store. Removing a key that does not exist is
// not an error:
//
//	remove.TheKey("app/config/cache")
func TheKey(key string) *TheKeyAction {
	return &TheKeyAction{key: key}
}

// TheKeyAction removes a key.
type TheKeyAction struct {
	key string
}

// String describes the action.
func (a *TheKeyAction) String() string {
	return "remove the key " + verdict.Describe(a.key)
}

// PerformAs removes the key.
func (a *TheKeyAction) PerformAs(theActor *screenplay.Actor) error {
	if a.key == "" {
		return action.ErrKeyNotProvided
	}

	return impl.TheAbilityOf(theActor).ToRemove(theActor.Context(), a.key)
}

// TheKeyAction implements the screenplay.Performable interface.
var _ screenplay.Performable = (*TheKeyAction)(nil)
