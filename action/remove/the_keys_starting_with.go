package remove

import (
	"github.com/grandper/go-screenplay/v2/screenplay"

	"github.com/grandper/go-screenplay-etcd/etcd"
	impl "github.com/grandper/go-screenplay-etcd/internal/ability"
	"github.com/grandper/go-screenplay-etcd/internal/verdict"
)

// TheKeysStartingWith removes every key starting with the prefix. Removing a
// prefix no key starts with is not an error:
//
//	remove.TheKeysStartingWith("app/config/")
func TheKeysStartingWith(prefix string) screenplay.Performable {
	return &theKeysStartingWithAction{prefix: prefix}
}

// theKeysStartingWithAction removes every key starting with a prefix.
type theKeysStartingWithAction struct {
	prefix string
}

// String describes the action.
func (a *theKeysStartingWithAction) String() string {
	return "remove the keys starting with " + verdict.Describe(a.prefix)
}

// PerformAs removes the keys starting with the prefix.
func (a *theKeysStartingWithAction) PerformAs(theActor *screenplay.Actor) error {
	if a.prefix == "" {
		return etcd.ErrPrefixNotProvided
	}

	return impl.TheAbilityOf(theActor).ToRemoveTheKeysStartingWith(theActor.Context(), a.prefix)
}

// theKeysStartingWithAction implements the screenplay.Performable interface.
var _ screenplay.Performable = (*theKeysStartingWithAction)(nil)
