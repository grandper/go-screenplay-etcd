package remove

import (
	"fmt"

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
	key         string
	conditional bool
	revision    int64
}

// IfUnchangedSinceTheRevision removes the key only when it was not modified
// since the revision, the one question.RevisionOfTheKey answered with; it
// reads "if unchanged since the revision 42". When the key was modified, the
// action fails with an error wrapping action.ErrKeyModified and the store is
// left untouched.
func (a *TheKeyAction) IfUnchangedSinceTheRevision(revision int64) *TheKeyAction {
	a.conditional = true
	a.revision = revision

	return a
}

// String describes the action.
func (a *TheKeyAction) String() string {
	description := "remove the key " + verdict.Describe(a.key)
	if a.conditional {
		description += fmt.Sprintf(" if unchanged since the revision %d", a.revision)
	}

	return description
}

// PerformAs removes the key.
func (a *TheKeyAction) PerformAs(theActor *screenplay.Actor) error {
	if a.key == "" {
		return action.ErrKeyNotProvided
	}

	theAbility := impl.TheAbilityOf(theActor)

	if !a.conditional {
		return theAbility.ToRemove(theActor.Context(), a.key)
	}

	succeeded, err := theAbility.ToRemoveIfUnchanged(theActor.Context(), a.key, a.revision)
	if err != nil {
		return err
	}
	if !succeeded {
		return fmt.Errorf(
			"the key %s %w since the revision %d",
			verdict.Describe(a.key),
			action.ErrKeyModified,
			a.revision,
		)
	}

	return nil
}

// TheKeyAction implements the screenplay.Performable interface.
var _ screenplay.Performable = (*TheKeyAction)(nil)
