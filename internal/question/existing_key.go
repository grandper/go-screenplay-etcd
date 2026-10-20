package question

import (
	"fmt"

	"github.com/grandper/go-screenplay/v2/screenplay"

	"github.com/grandper/go-screenplay-etcd/ability"
	impl "github.com/grandper/go-screenplay-etcd/internal/ability"
	"github.com/grandper/go-screenplay-etcd/internal/verdict"
	"github.com/grandper/go-screenplay-etcd/question"
)

// RetrieveTheExistingKey retrieves the key for the questions that need it to
// exist: an empty key fails with question.ErrKeyNotProvided, and a key that
// does not exist with an error wrapping question.ErrKeyNotFound.
func RetrieveTheExistingKey(theActor *screenplay.Actor, name string) (*ability.Key, error) {
	if name == "" {
		return nil, question.ErrKeyNotProvided
	}

	key, err := impl.TheAbilityOf(theActor).ToRetrieveTheKey(theActor.Context(), name)
	if err != nil {
		return nil, err
	}

	if !key.Exists() {
		return nil, fmt.Errorf("the key %s %w", verdict.Describe(name), question.ErrKeyNotFound)
	}

	return key, nil
}
