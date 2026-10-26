package values

import (
	"github.com/grandper/go-screenplay/v2/screenplay"

	"github.com/grandper/go-screenplay-etcd/etcd"

	impl "github.com/grandper/go-screenplay-etcd/internal/ability"
	"github.com/grandper/go-screenplay-etcd/internal/verdict"
)

// OfTheKeysStartingWith asks about the values of the keys starting with
// a prefix, as a map from each key to its value. When no key starts with the
// prefix, the map is empty:
//
//	err := anActor.Should(see.The(values.OfTheKeysStartingWith("app/config/")).Should(be.Empty[map[string]string]()))
func OfTheKeysStartingWith(prefix string) screenplay.Question[map[string]string] {
	return &ofTheKeysStartingWithQuestion{prefix: prefix}
}

// ofTheKeysStartingWithQuestion asks about the values of the keys
// starting with a prefix.
type ofTheKeysStartingWithQuestion struct {
	prefix string
}

// AnsweredBy returns the values of the keys starting with the prefix.
func (q *ofTheKeysStartingWithQuestion) AnsweredBy(theActor *screenplay.Actor) (map[string]string, error) {
	if q.prefix == "" {
		return nil, etcd.ErrPrefixNotProvided
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
func (q *ofTheKeysStartingWithQuestion) String() string {
	return "values of the keys starting with " + verdict.Describe(q.prefix)
}

// ofTheKeysStartingWithQuestion implements the screenplay.Question interface.
var _ screenplay.Question[map[string]string] = (*ofTheKeysStartingWithQuestion)(nil)
