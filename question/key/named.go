package key

import (
	"github.com/grandper/go-screenplay/v2/screenplay"

	"github.com/grandper/go-screenplay-etcd/question"

	"github.com/grandper/go-screenplay-etcd/ability"
	impl "github.com/grandper/go-screenplay-etcd/internal/ability"
	"github.com/grandper/go-screenplay-etcd/internal/verdict"
)

// Named asks about a key, whether it exists or not: its value and its
// metadata are read through the *ability.Key it answers with.
//
//	err := anActor.Should(see.The(key.Named("app/status")).Should(exist.InETCD()))
func Named(key string) screenplay.Question[*ability.Key] {
	return &namedQuestion{key: key}
}

// namedQuestion asks about a key, whether it exists or not.
type namedQuestion struct {
	key string
}

// AnsweredBy returns the key, which exists or not.
func (q *namedQuestion) AnsweredBy(theActor *screenplay.Actor) (*ability.Key, error) {
	if q.key == "" {
		return nil, question.ErrKeyNotProvided
	}

	return impl.TheAbilityOf(theActor).ToRetrieveTheKey(theActor.Context(), q.key)
}

// String describes the question.
func (q *namedQuestion) String() string {
	return "key named " + verdict.Describe(q.key)
}

// namedQuestion implements the screenplay.Question interface.
var _ screenplay.Question[*ability.Key] = (*namedQuestion)(nil)
