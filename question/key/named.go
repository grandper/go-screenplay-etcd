package key

import (
	"github.com/grandper/go-screenplay/v2/screenplay"

	"github.com/grandper/go-screenplay-etcd/etcd"

	impl "github.com/grandper/go-screenplay-etcd/internal/ability"
	"github.com/grandper/go-screenplay-etcd/internal/verdict"
)

// Named asks about a key, whether it exists or not: its value and its
// metadata are read through the *etcd.Key it answers with.
//
//	err := anActor.Should(see.The(key.Named("app/status")).Should(exist.InETCD()))
func Named(key string) screenplay.Question[*etcd.Key] {
	return &namedQuestion{key: key}
}

// namedQuestion asks about a key, whether it exists or not.
type namedQuestion struct {
	key string
}

// AnsweredBy returns the key, which exists or not.
func (q *namedQuestion) AnsweredBy(theActor *screenplay.Actor) (*etcd.Key, error) {
	if q.key == "" {
		return nil, etcd.ErrKeyNotProvided
	}

	return impl.TheAbilityOf(theActor).ToRetrieveTheKey(theActor.Context(), q.key)
}

// String describes the question.
func (q *namedQuestion) String() string {
	return "key named " + verdict.Describe(q.key)
}

// namedQuestion implements the screenplay.Question interface.
var _ screenplay.Question[*etcd.Key] = (*namedQuestion)(nil)
