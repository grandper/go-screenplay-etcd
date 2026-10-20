package campaign

import (
	"github.com/grandper/go-screenplay/v2/screenplay"

	"github.com/grandper/go-screenplay-etcd/action"
	impl "github.com/grandper/go-screenplay-etcd/internal/ability"
	"github.com/grandper/go-screenplay-etcd/internal/verdict"
)

// InTheElection campaigns in an election, waiting for the leader to resign if
// there is one, and returns once the actor leads it. The leadership is held
// until the actor resigns with resign.FromTheElection, or exits. The leader
// is announced under the name of the actor, unless a value is given with
// WithTheValue:
//
//	campaign.InTheElection("primary-worker")
func InTheElection(name string) *InTheElectionAction {
	return &InTheElectionAction{name: name}
}

// InTheElectionAction campaigns in an election.
type InTheElectionAction struct {
	name     string
	value    string
	hasValue bool
}

// WithTheValue sets the value the leader is announced with, instead of the
// name of the actor; it reads "with the value node-1".
func (a *InTheElectionAction) WithTheValue(value string) *InTheElectionAction {
	a.value = value
	a.hasValue = true

	return a
}

// String describes the action.
func (a *InTheElectionAction) String() string {
	description := "campaign in the election " + verdict.Describe(a.name)
	if a.hasValue {
		description += " with the value " + verdict.Describe(a.value)
	}

	return description
}

// PerformAs campaigns in the election.
func (a *InTheElectionAction) PerformAs(theActor *screenplay.Actor) error {
	if a.name == "" {
		return action.ErrNameNotProvided
	}

	value := a.value
	if !a.hasValue {
		value = theActor.Name()
	}

	return impl.TheAbilityOf(theActor).ToCampaign(theActor.Context(), a.name, value)
}

// InTheElectionAction implements the screenplay.Performable interface.
var _ screenplay.Performable = (*InTheElectionAction)(nil)
