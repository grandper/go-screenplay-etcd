package remove

import (
	"strings"

	"github.com/grandper/go-screenplay/v2/screenplay"

	"github.com/grandper/go-screenplay-etcd/action"
	impl "github.com/grandper/go-screenplay-etcd/internal/ability"
	"github.com/grandper/go-screenplay-etcd/internal/verdict"
)

// TheKeys removes several keys from the store in one transaction, so that the
// store never holds some of them without the others. Removing a key that does
// not exist is not an error:
//
//	remove.TheKeys("app/config/cache", "app/config/database")
func TheKeys(keys ...string) screenplay.Performable {
	return &theKeysAction{keys: keys}
}

// theKeysAction removes several keys in one transaction.
type theKeysAction struct {
	keys []string
}

// String describes the action.
func (a *theKeysAction) String() string {
	if len(a.keys) == 0 {
		return "remove the keys"
	}

	described := make([]string, 0, len(a.keys))
	for _, key := range a.keys {
		described = append(described, verdict.Describe(key))
	}

	return "remove the keys " + strings.Join(described, ", ")
}

// PerformAs removes the keys.
func (a *theKeysAction) PerformAs(theActor *screenplay.Actor) error {
	if len(a.keys) == 0 {
		return action.ErrKeyNotProvided
	}

	for _, key := range a.keys {
		if key == "" {
			return action.ErrKeyNotProvided
		}
	}

	return impl.TheAbilityOf(theActor).ToRemoveTheKeys(theActor.Context(), a.keys...)
}

// theKeysAction implements the screenplay.Performable interface.
var _ screenplay.Performable = (*theKeysAction)(nil)
