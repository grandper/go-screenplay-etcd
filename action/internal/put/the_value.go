package put

import (
	"fmt"
	"time"

	"github.com/grandper/go-screenplay/v2/action/timing"
	"github.com/grandper/go-screenplay/v2/screenplay"
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/grandper/go-screenplay-etcd/action"
	impl "github.com/grandper/go-screenplay-etcd/internal/ability"
	"github.com/grandper/go-screenplay-etcd/internal/verdict"
)

// TheValue puts a value in the store. The key is given with ForTheKey, so the
// statement reads as a sentence:
//
//	put.TheValue("postgresql://localhost:5432").ForTheKey("app/config/database")
func TheValue(value string) *TheValueAction {
	a := &TheValueAction{value: value}
	a.expiry = timing.NewDurationBuilder(a, &a.timeToLive)

	return a
}

// TheValueAction puts a value for a key.
type TheValueAction struct {
	value  string
	key    string
	secret bool
	// condition guards the put: a put that is not conditional, that requires
	// the key to be absent, or that requires the key to be unchanged since a
	// revision.
	condition condition
	revision  int64
	// expiry builds the time to live of the key, written to timeToLive; zero
	// means that the key never expires.
	expiry     *timing.DurationBuilder[TheValueAction]
	timeToLive time.Duration
	keptAlive  bool
}

// condition is the condition a put is guarded by.
type condition int

const (
	unconditional condition = iota
	ifTheKeyIsAbsent
	ifTheKeyIsUnchanged
)

// ForTheKey sets the key the value is put for.
func (a *TheValueAction) ForTheKey(key string) *TheValueAction {
	a.key = key

	return a
}

// IfTheKeyIsAbsent puts the value only when the key does not exist; it reads
// "if the key is absent". When the key exists, the action fails with an error
// wrapping action.ErrKeyAlreadyExists and the store is left untouched.
func (a *TheValueAction) IfTheKeyIsAbsent() *TheValueAction {
	a.condition = ifTheKeyIsAbsent

	return a
}

// IfTheKeyIsUnchangedSinceTheRevision puts the value only when the key was not
// modified since the revision, the one question.RevisionOfTheKey answered
// with; it reads "if the key is unchanged since the revision 42". The revision
// zero means that the key must not exist. When the key was modified, the
// action fails with an error wrapping action.ErrKeyModified and the store is
// left untouched.
func (a *TheValueAction) IfTheKeyIsUnchangedSinceTheRevision(revision int64) *TheValueAction {
	a.condition = ifTheKeyIsUnchanged
	a.revision = revision

	return a
}

// ExpiringAfter attaches the key to a lease living for the given time, so
// that the key is removed when the lease expires; it reads "expiring after 10
// seconds". The time is closed by its unit, and rounded up to whole seconds,
// which is how etcd counts.
func (a *TheValueAction) ExpiringAfter(amount int) *timing.DurationBuilder[TheValueAction] {
	return a.expiry.For(amount)
}

// AndKeepItAlive renews the lease of the key in the background for as long as
// the actor keeps the ability, so that the key expires only once the actor
// exits; it reads "and keep it alive".
func (a *TheValueAction) AndKeepItAlive() *TheValueAction {
	a.keptAlive = true

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

	if a.timeToLive > 0 {
		description += " expiring after " + a.expiry.String()
	}

	if a.keptAlive {
		description += " and keep it alive"
	}

	switch a.condition {
	case ifTheKeyIsAbsent:
		description += " if the key is absent"
	case ifTheKeyIsUnchanged:
		description += fmt.Sprintf(" if the key is unchanged since the revision %d", a.revision)
	case unconditional:
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

	theAbility := impl.TheAbilityOf(theActor)

	lease, err := a.grantTheLease(theActor, theAbility)
	if err != nil {
		return err
	}

	if err = a.put(theActor, theAbility, lease); err != nil {
		return err
	}

	if a.keptAlive && lease != clientv3.NoLease {
		return theAbility.ToKeepTheLeaseAlive(lease)
	}

	return nil
}

// grantTheLease grants the lease the key is attached to, or no lease when the
// key never expires.
func (a *TheValueAction) grantTheLease(
	theActor *screenplay.Actor,
	theAbility *impl.Operations,
) (clientv3.LeaseID, error) {
	if a.timeToLive <= 0 {
		return clientv3.NoLease, nil
	}

	return theAbility.ToGrantALease(theActor.Context(), a.timeToLive)
}

// put puts the value for the key, under the condition of the action.
func (a *TheValueAction) put(theActor *screenplay.Actor, theAbility *impl.Operations, lease clientv3.LeaseID) error {
	switch a.condition {
	case ifTheKeyIsAbsent:
		succeeded, err := theAbility.ToPutIfAbsent(theActor.Context(), a.key, a.value, lease)
		if err != nil {
			return err
		}
		if !succeeded {
			return fmt.Errorf("the key %s %w", verdict.Describe(a.key), action.ErrKeyAlreadyExists)
		}

		return nil
	case ifTheKeyIsUnchanged:
		succeeded, err := theAbility.ToPutIfUnchanged(theActor.Context(), a.key, a.value, a.revision, lease)
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
	case unconditional:
	}

	return theAbility.ToPut(theActor.Context(), a.key, a.value, lease)
}

// TheValueAction implements the screenplay.Performable interface.
var _ screenplay.Performable = (*TheValueAction)(nil)
