package keys

import (
	"github.com/grandper/go-screenplay/v2/screenplay"

	"github.com/grandper/go-screenplay-etcd/etcd"

	impl "github.com/grandper/go-screenplay-etcd/internal/ability"
	"github.com/grandper/go-screenplay-etcd/internal/verdict"
)

// StartingWith asks about the keys starting with a prefix, in the order of
// the keys. When no key starts with the prefix, the slice is empty:
//
//	err := anActor.Should(see.The(keys.StartingWith("app/config/")).ShouldHave(length.Exactly[[]string](2)))
func StartingWith(prefix string) screenplay.Question[[]string] {
	return &startingWithQuestion{prefix: prefix}
}

// startingWithQuestion asks about the keys starting with a prefix.
type startingWithQuestion struct {
	prefix string
}

// AnsweredBy returns the keys starting with the prefix.
func (q *startingWithQuestion) AnsweredBy(theActor *screenplay.Actor) ([]string, error) {
	if q.prefix == "" {
		return nil, etcd.ErrPrefixNotProvided
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
func (q *startingWithQuestion) String() string {
	return "keys starting with " + verdict.Describe(q.prefix)
}

// startingWithQuestion implements the screenplay.Question interface.
var _ screenplay.Question[[]string] = (*startingWithQuestion)(nil)
