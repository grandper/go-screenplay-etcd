package question

import (
	"github.com/grandper/go-screenplay/v2/screenplay"

	impl "github.com/grandper/go-screenplay-etcd/internal/ability"
	"github.com/grandper/go-screenplay-etcd/internal/verdict"
)

// ValuesOfTheKeysStartingWith asks about the values of the keys starting with
// a prefix, as a map from each key to its value. When no key starts with the
// prefix, the map is empty:
//
//	err := anActor.Should(see.The(question.ValuesOfTheKeysStartingWith("app/config/")).Should(be.Empty[map[string]string]()))
func ValuesOfTheKeysStartingWith(prefix string) screenplay.Question[map[string]string] {
	return &valuesOfTheKeysStartingWithQuestion{prefix: prefix}
}

// valuesOfTheKeysStartingWithQuestion asks about the values of the keys
// starting with a prefix.
type valuesOfTheKeysStartingWithQuestion struct {
	prefix string
}

// AnsweredBy returns the values of the keys starting with the prefix.
func (q *valuesOfTheKeysStartingWithQuestion) AnsweredBy(theActor *screenplay.Actor) (map[string]string, error) {
	if q.prefix == "" {
		return nil, ErrPrefixNotProvided
	}

	keys, err := impl.TheAbilityOf(theActor).ToRetrieveTheKeysStartingWith(theActor.Context(), q.prefix)
	if err != nil {
		return nil, err
	}

	values := make(map[string]string, len(keys))
	for _, key := range keys {
		values[key.Name()] = key.Value()
	}

	return values, nil
}

// String describes the question.
func (q *valuesOfTheKeysStartingWithQuestion) String() string {
	return "values of the keys starting with " + verdict.Describe(q.prefix)
}

// valuesOfTheKeysStartingWithQuestion implements the screenplay.Question interface.
var _ screenplay.Question[map[string]string] = (*valuesOfTheKeysStartingWithQuestion)(nil)
