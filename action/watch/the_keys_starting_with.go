package watch

import (
	"github.com/grandper/go-screenplay/v2/screenplay"

	"github.com/grandper/go-screenplay-etcd/action"
	impl "github.com/grandper/go-screenplay-etcd/internal/ability"
	"github.com/grandper/go-screenplay-etcd/internal/verdict"
)

// TheKeysStartingWith watches the keys starting with a prefix, recording every
// change made to any of them until the actor exits:
//
//	watch.TheKeysStartingWith("app/config/")
func TheKeysStartingWith(prefix string) screenplay.Performable {
	return &theKeysStartingWithAction{prefix: prefix}
}

// theKeysStartingWithAction watches the keys starting with a prefix.
type theKeysStartingWithAction struct {
	prefix string
}

// String describes the action.
func (a *theKeysStartingWithAction) String() string {
	return "watch the keys starting with " + verdict.Describe(a.prefix)
}

// PerformAs watches the keys starting with the prefix.
func (a *theKeysStartingWithAction) PerformAs(theActor *screenplay.Actor) error {
	if a.prefix == "" {
		return action.ErrPrefixNotProvided
	}

	return impl.TheAbilityOf(theActor).ToWatch(theActor.Context(), a.prefix, true)
}

// theKeysStartingWithAction implements the screenplay.Performable interface.
var _ screenplay.Performable = (*theKeysStartingWithAction)(nil)
