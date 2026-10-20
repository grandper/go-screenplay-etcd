package question

import (
	"github.com/grandper/go-screenplay/v2/screenplay"

	impl "github.com/grandper/go-screenplay-etcd/internal/ability"
	"github.com/grandper/go-screenplay-etcd/internal/verdict"
)

// KeysStartingWith asks about the keys starting with a prefix, in the order of
// the keys. When no key starts with the prefix, the slice is empty:
//
//	err := anActor.Should(see.The(question.KeysStartingWith("app/config/")).ShouldHave(length.Exactly[[]string](2)))
func KeysStartingWith(prefix string) screenplay.Question[[]string] {
	return &keysStartingWithQuestion{prefix: prefix}
}

// keysStartingWithQuestion asks about the keys starting with a prefix.
type keysStartingWithQuestion struct {
	prefix string
}

// AnsweredBy returns the keys starting with the prefix.
func (q *keysStartingWithQuestion) AnsweredBy(theActor *screenplay.Actor) ([]string, error) {
	if q.prefix == "" {
		return nil, ErrPrefixNotProvided
	}

	keys, err := impl.TheAbilityOf(theActor).ToRetrieveTheKeysStartingWith(theActor.Context(), q.prefix)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(keys))
	for _, key := range keys {
		names = append(names, key.Name())
	}

	return names, nil
}

// String describes the question.
func (q *keysStartingWithQuestion) String() string {
	return "keys starting with " + verdict.Describe(q.prefix)
}

// keysStartingWithQuestion implements the screenplay.Question interface.
var _ screenplay.Question[[]string] = (*keysStartingWithQuestion)(nil)
