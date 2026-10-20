package put

import (
	"github.com/grandper/go-screenplay/v2/screenplay"

	"github.com/grandper/go-screenplay-etcd/action"
	impl "github.com/grandper/go-screenplay-etcd/internal/ability"
	"github.com/grandper/go-screenplay-etcd/internal/verdict"
)

// TheValue puts a value in the store. The key is given with ForTheKey, so the
// statement reads as a sentence:
//
//	put.TheValue("postgresql://localhost:5432").ForTheKey("app/config/database")
func TheValue(value string) *TheValueAction {
	return &TheValueAction{value: value}
}

// TheValueAction puts a value for a key.
type TheValueAction struct {
	value  string
	key    string
	secret bool
}

// ForTheKey sets the key the value is put for.
func (a *TheValueAction) ForTheKey(key string) *TheValueAction {
	a.key = key

	return a
}

// WhichShouldBeKeptSecret makes sure that the value is not displayed in the
// narration.
func (a *TheValueAction) WhichShouldBeKeptSecret() *TheValueAction {
	a.secret = true

	return a
}

// Secretly makes sure that the value is not displayed in the narration.
// Secretly is an alias for the method WhichShouldBeKeptSecret.
func (a *TheValueAction) Secretly() *TheValueAction {
	return a.WhichShouldBeKeptSecret()
}

// String describes the action.
func (a *TheValueAction) String() string {
	description := "put the value " + a.describeTheValue()
	if a.key != "" {
		description += " for the key " + verdict.Describe(a.key)
	}

	return description
}

// describeTheValue describes the value, or hides it when it is secret.
func (a *TheValueAction) describeTheValue() string {
	if a.secret {
		return "<secret>"
	}

	return verdict.Describe(a.value)
}

// PerformAs puts the value for the key.
func (a *TheValueAction) PerformAs(theActor *screenplay.Actor) error {
	if a.key == "" {
		return action.ErrKeyNotProvided
	}

	return impl.TheAbilityOf(theActor).ToPut(theActor.Context(), a.key, a.value)
}

// TheValueAction implements the screenplay.Performable interface.
var _ screenplay.Performable = (*TheValueAction)(nil)
