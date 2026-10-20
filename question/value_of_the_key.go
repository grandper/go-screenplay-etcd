package question

import (
	"fmt"

	"github.com/grandper/go-screenplay/v2/screenplay"

	impl "github.com/grandper/go-screenplay-etcd/internal/ability"
	"github.com/grandper/go-screenplay-etcd/internal/verdict"
)

// ValueOfTheKey asks about the value a key holds:
//
//	err := anActor.Should(see.The(question.ValueOfTheKey("app/config/database")).Should(be.EqualTo("postgresql://localhost:5432")))
//
// The question fails with an error wrapping ErrKeyNotFound when the key does
// not exist.
func ValueOfTheKey(key string) screenplay.Question[string] {
	return &valueOfTheKeyQuestion{key: key}
}

// valueOfTheKeyQuestion asks about the value a key holds.
type valueOfTheKeyQuestion struct {
	key string
}

// AnsweredBy returns the value the key holds.
func (q *valueOfTheKeyQuestion) AnsweredBy(theActor *screenplay.Actor) (string, error) {
	key, err := impl.TheAbilityOf(theActor).ToRetrieveTheKey(theActor.Context(), q.key)
	if err != nil {
		return "", err
	}

	if !key.Exists() {
		return "", fmt.Errorf("the key %s %w", verdict.Describe(q.key), ErrKeyNotFound)
	}

	return key.Value(), nil
}

// String describes the question.
func (q *valueOfTheKeyQuestion) String() string {
	return "value of the key " + verdict.Describe(q.key)
}

// valueOfTheKeyQuestion implements the screenplay.Question interface.
var _ screenplay.Question[string] = (*valueOfTheKeyQuestion)(nil)
