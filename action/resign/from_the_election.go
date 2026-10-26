package resign

import (
	"github.com/grandper/go-screenplay/v2/screenplay"

	"github.com/grandper/go-screenplay-etcd/etcd"
	impl "github.com/grandper/go-screenplay-etcd/internal/ability"
	"github.com/grandper/go-screenplay-etcd/internal/verdict"
)

// FromTheElection resigns from an election the actor campaigns in, giving up
// its leadership:
//
//	resign.FromTheElection("primary-worker")
func FromTheElection(name string) screenplay.Performable {
	return &fromTheElectionAction{name: name}
}

// fromTheElectionAction resigns from an election.
type fromTheElectionAction struct {
	name string
}

// String describes the action.
func (a *fromTheElectionAction) String() string {
	return "resign from the election " + verdict.Describe(a.name)
}

// PerformAs resigns from the election.
func (a *fromTheElectionAction) PerformAs(theActor *screenplay.Actor) error {
	if a.name == "" {
		return etcd.ErrNameNotProvided
	}

	return impl.TheAbilityOf(theActor).ToResign(theActor.Context(), a.name)
}

// fromTheElectionAction implements the screenplay.Performable interface.
var _ screenplay.Performable = (*fromTheElectionAction)(nil)
