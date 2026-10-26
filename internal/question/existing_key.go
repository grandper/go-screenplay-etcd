package question

import (
	"fmt"

	"github.com/grandper/go-screenplay/v2/screenplay"

	"github.com/grandper/go-screenplay-etcd/etcd"
	impl "github.com/grandper/go-screenplay-etcd/internal/ability"
	"github.com/grandper/go-screenplay-etcd/internal/verdict"
)

// RetrieveTheExistingKey retrieves the key for the questions that need it to
// exist: an empty key fails with etcd.ErrKeyNotProvided, and a key that
// does not exist with an error wrapping etcd.ErrKeyNotFound.
func RetrieveTheExistingKey(theActor *screenplay.Actor, name string) (*etcd.Key, error) {
	if name == "" {
		return nil, etcd.ErrKeyNotProvided
	}

	key, err := impl.TheAbilityOf(theActor).ToRetrieveTheKey(theActor.Context(), name)
	if err != nil {
		return nil, err
	}

	if !key.Exists() {
		return nil, fmt.Errorf("the key %s %w", verdict.Describe(name), etcd.ErrKeyNotFound)
	}

	return key, nil
}
