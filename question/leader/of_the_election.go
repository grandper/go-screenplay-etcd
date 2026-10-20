package leader

import (
	"github.com/grandper/go-screenplay/v2/screenplay"

	"github.com/grandper/go-screenplay-etcd/question"

	impl "github.com/grandper/go-screenplay-etcd/internal/ability"
	"github.com/grandper/go-screenplay-etcd/internal/verdict"
)

// OfTheElection asks about the value the leader of an election was
// announced with, the name of the actor leading it unless it campaigned with
// a value of its own. An election without a leader answers with an empty
// value:
//
//	err := anActor.Should(see.The(leader.OfTheElection("primary-worker")).Should(be.EqualTo("Alice")))
func OfTheElection(name string) screenplay.Question[string] {
	return &ofTheElectionQuestion{name: name}
}

// ofTheElectionQuestion asks about the leader of an election.
type ofTheElectionQuestion struct {
	name string
}

// AnsweredBy returns the value of the leader of the election.
func (q *ofTheElectionQuestion) AnsweredBy(theActor *screenplay.Actor) (string, error) {
	if q.name == "" {
		return "", question.ErrNameNotProvided
	}

	return impl.TheAbilityOf(theActor).ToRetrieveTheLeader(theActor.Context(), q.name)
}

// String describes the question.
func (q *ofTheElectionQuestion) String() string {
	return "leader of the election " + verdict.Describe(q.name)
}

// ofTheElectionQuestion implements the screenplay.Question interface.
var _ screenplay.Question[string] = (*ofTheElectionQuestion)(nil)
