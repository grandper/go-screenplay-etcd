package question

import (
	"fmt"

	"github.com/grandper/go-screenplay/v2/screenplay"

	"github.com/grandper/go-screenplay-etcd/ability"
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
	key, err := retrieveTheExistingKey(theActor, q.key)
	if err != nil {
		return "", err
	}

	return key.Value(), nil
}

// String describes the question.
func (q *valueOfTheKeyQuestion) String() string {
	return "value of the key " + verdict.Describe(q.key)
}

// valueOfTheKeyQuestion implements the screenplay.Question interface.
var _ screenplay.Question[string] = (*valueOfTheKeyQuestion)(nil)

// retrieveTheExistingKey retrieves the key for the questions that need it to
// exist: an empty key fails with ErrKeyNotProvided, and a key that does not
// exist with an error wrapping ErrKeyNotFound.
func retrieveTheExistingKey(theActor *screenplay.Actor, name string) (*ability.Key, error) {
	if name == "" {
		return nil, ErrKeyNotProvided
	}

	key, err := impl.TheAbilityOf(theActor).ToRetrieveTheKey(theActor.Context(), name)
	if err != nil {
		return nil, err
	}

	if !key.Exists() {
		return nil, fmt.Errorf("the key %s %w", verdict.Describe(name), ErrKeyNotFound)
	}

	return key, nil
}
