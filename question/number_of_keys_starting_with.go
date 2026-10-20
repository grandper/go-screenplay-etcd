package question

import (
	"github.com/grandper/go-screenplay/v2/screenplay"

	impl "github.com/grandper/go-screenplay-etcd/internal/ability"
	"github.com/grandper/go-screenplay-etcd/internal/verdict"
)

// NumberOfKeysStartingWith asks about the number of keys starting with a
// prefix, counted by the store without retrieving them:
//
//	err := anActor.Should(see.The(question.NumberOfKeysStartingWith("app/config/")).Should(be.EqualTo(2)))
func NumberOfKeysStartingWith(prefix string) screenplay.Question[int] {
	return &numberOfKeysStartingWithQuestion{prefix: prefix}
}

// numberOfKeysStartingWithQuestion asks about the number of keys starting with
// a prefix.
type numberOfKeysStartingWithQuestion struct {
	prefix string
}

// AnsweredBy returns the number of keys starting with the prefix.
func (q *numberOfKeysStartingWithQuestion) AnsweredBy(theActor *screenplay.Actor) (int, error) {
	if q.prefix == "" {
		return 0, ErrPrefixNotProvided
	}

	return impl.TheAbilityOf(theActor).ToCountTheKeysStartingWith(theActor.Context(), q.prefix)
}

// String describes the question.
func (q *numberOfKeysStartingWithQuestion) String() string {
	return "number of keys starting with " + verdict.Describe(q.prefix)
}

// numberOfKeysStartingWithQuestion implements the screenplay.Question interface.
var _ screenplay.Question[int] = (*numberOfKeysStartingWithQuestion)(nil)
