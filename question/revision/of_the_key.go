package revision

import (
	"github.com/grandper/go-screenplay/v2/screenplay"

	internalquestion "github.com/grandper/go-screenplay-etcd/internal/question"

	"github.com/grandper/go-screenplay-etcd/internal/verdict"
)

// OfTheKey asks about the revision of the store at which a key was
// last modified, the one to give to a conditional put or remove:
//
//	revision, err := anActor.AsksFor(revision.OfTheKey("app/config/database"))
//
// The question fails with an error wrapping etcd.ErrKeyNotFound when the key does
// not exist.
func OfTheKey(key string) screenplay.Question[int64] {
	return &ofTheKeyQuestion{key: key}
}

// ofTheKeyQuestion asks about the revision at which a key was last
// modified.
type ofTheKeyQuestion struct {
	key string
}

// AnsweredBy returns the revision at which the key was last modified.
func (q *ofTheKeyQuestion) AnsweredBy(theActor *screenplay.Actor) (int64, error) {
	key, err := internalquestion.RetrieveTheExistingKey(theActor, q.key)
	if err != nil {
		return 0, err
	}

	return key.ModRevision(), nil
}

// String describes the question.
func (q *ofTheKeyQuestion) String() string {
	return "revision of the key " + verdict.Describe(q.key)
}

// ofTheKeyQuestion implements the screenplay.Question interface.
var _ screenplay.Question[int64] = (*ofTheKeyQuestion)(nil)
