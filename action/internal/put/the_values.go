package put

import (
	"maps"
	"slices"
	"strings"

	"github.com/grandper/go-screenplay/v2/screenplay"

	"github.com/grandper/go-screenplay-etcd/action"
	impl "github.com/grandper/go-screenplay-etcd/internal/ability"
	"github.com/grandper/go-screenplay-etcd/internal/verdict"
)

// TheValues puts several values in the store, each for its key, in one
// transaction, so that the store never holds some of them without the others:
//
//	put.TheValues(map[string]string{
//		"app/config/database": "postgresql://localhost:5432",
//		"app/config/cache":    "redis://localhost:6379",
//	})
func TheValues(values map[string]string) *TheValuesAction {
	return &TheValuesAction{values: values}
}

// TheValuesAction puts several values, each for its key, in one transaction.
type TheValuesAction struct {
	values map[string]string
	secret bool
}

// WhichShouldBeKeptSecret makes sure that the values are not displayed in the
// narration.
func (a *TheValuesAction) WhichShouldBeKeptSecret() *TheValuesAction {
	a.secret = true

	return a
}

// Secretly makes sure that the values are not displayed in the narration.
// Secretly is an alias for the method WhichShouldBeKeptSecret.
func (a *TheValuesAction) Secretly() *TheValuesAction {
	return a.WhichShouldBeKeptSecret()
}

// String describes the action, with the values sorted by key.
func (a *TheValuesAction) String() string {
	if len(a.values) == 0 {
		return "put the values"
	}

	keys := slices.Sorted(maps.Keys(a.values))

	if a.secret {
		described := make([]string, 0, len(keys))
		for _, key := range keys {
			described = append(described, verdict.Describe(key))
		}

		return "put the values <secret> for the keys " + strings.Join(described, ", ")
	}

	described := make([]string, 0, len(keys))
	for _, key := range keys {
		described = append(described, verdict.Describe(key)+" = "+verdict.Describe(a.values[key]))
	}

	return "put the values " + strings.Join(described, ", ")
}

// PerformAs puts the values.
func (a *TheValuesAction) PerformAs(theActor *screenplay.Actor) error {
	if len(a.values) == 0 {
		return action.ErrKeyNotProvided
	}

	if _, hasAnEmptyKey := a.values[""]; hasAnEmptyKey {
		return action.ErrKeyNotProvided
	}

	return impl.TheAbilityOf(theActor).ToPutTheValues(theActor.Context(), a.values)
}

// TheValuesAction implements the screenplay.Performable interface.
var _ screenplay.Performable = (*TheValuesAction)(nil)
