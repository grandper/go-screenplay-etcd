package question

import (
	"github.com/grandper/go-screenplay/v2/screenplay"

	"github.com/grandper/go-screenplay-etcd/ability"
	impl "github.com/grandper/go-screenplay-etcd/internal/ability"
	"github.com/grandper/go-screenplay-etcd/internal/verdict"
)

// KeyNamed asks about a key, whether it exists or not: its value and its
// metadata are read through the *ability.Key it answers with.
//
//	err := anActor.Should(see.The(question.KeyNamed("app/status")).Should(resolution.Exist()))
func KeyNamed(key string) screenplay.Question[*ability.Key] {
	return &keyNamedQuestion{key: key}
}

// keyNamedQuestion asks about a key, whether it exists or not.
type keyNamedQuestion struct {
	key string
}

// AnsweredBy returns the key, which exists or not.
func (q *keyNamedQuestion) AnsweredBy(theActor *screenplay.Actor) (*ability.Key, error) {
	if q.key == "" {
		return nil, ErrKeyNotProvided
	}

	return impl.TheAbilityOf(theActor).ToRetrieveTheKey(theActor.Context(), q.key)
}

// String describes the question.
func (q *keyNamedQuestion) String() string {
	return "key named " + verdict.Describe(q.key)
}

// keyNamedQuestion implements the screenplay.Question interface.
var _ screenplay.Question[*ability.Key] = (*keyNamedQuestion)(nil)
