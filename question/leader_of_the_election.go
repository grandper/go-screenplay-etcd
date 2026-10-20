package question

import (
	"github.com/grandper/go-screenplay/v2/screenplay"

	impl "github.com/grandper/go-screenplay-etcd/internal/ability"
	"github.com/grandper/go-screenplay-etcd/internal/verdict"
)

// LeaderOfTheElection asks about the value the leader of an election was
// announced with, the name of the actor leading it unless it campaigned with
// a value of its own. An election without a leader answers with an empty
// value:
//
//	err := anActor.Should(see.The(question.LeaderOfTheElection("primary-worker")).Should(be.EqualTo("Alice")))
func LeaderOfTheElection(name string) screenplay.Question[string] {
	return &leaderOfTheElectionQuestion{name: name}
}

// leaderOfTheElectionQuestion asks about the leader of an election.
type leaderOfTheElectionQuestion struct {
	name string
}

// AnsweredBy returns the value of the leader of the election.
func (q *leaderOfTheElectionQuestion) AnsweredBy(theActor *screenplay.Actor) (string, error) {
	if q.name == "" {
		return "", ErrNameNotProvided
	}

	return impl.TheAbilityOf(theActor).ToRetrieveTheLeader(theActor.Context(), q.name)
}

// String describes the question.
func (q *leaderOfTheElectionQuestion) String() string {
	return "leader of the election " + verdict.Describe(q.name)
}

// leaderOfTheElectionQuestion implements the screenplay.Question interface.
var _ screenplay.Question[string] = (*leaderOfTheElectionQuestion)(nil)
