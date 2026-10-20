package question

import (
	"github.com/grandper/go-screenplay/v2/screenplay"

	"github.com/grandper/go-screenplay-etcd/internal/verdict"
)

// RevisionOfTheKey asks about the revision of the store at which a key was
// last modified, the one to give to a conditional put or remove:
//
//	revision, err := anActor.AsksFor(question.RevisionOfTheKey("app/config/database"))
//
// The question fails with an error wrapping ErrKeyNotFound when the key does
// not exist.
func RevisionOfTheKey(key string) screenplay.Question[int64] {
	return &revisionOfTheKeyQuestion{key: key}
}

// revisionOfTheKeyQuestion asks about the revision at which a key was last
// modified.
type revisionOfTheKeyQuestion struct {
	key string
}

// AnsweredBy returns the revision at which the key was last modified.
func (q *revisionOfTheKeyQuestion) AnsweredBy(theActor *screenplay.Actor) (int64, error) {
	key, err := retrieveTheExistingKey(theActor, q.key)
	if err != nil {
		return 0, err
	}

	return key.ModRevision(), nil
}

// String describes the question.
func (q *revisionOfTheKeyQuestion) String() string {
	return "revision of the key " + verdict.Describe(q.key)
}

// revisionOfTheKeyQuestion implements the screenplay.Question interface.
var _ screenplay.Question[int64] = (*revisionOfTheKeyQuestion)(nil)
