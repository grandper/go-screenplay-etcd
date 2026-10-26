package value

import (
	internalquestion "github.com/grandper/go-screenplay-etcd/internal/question"

	"github.com/grandper/go-screenplay/v2/screenplay"

	"github.com/grandper/go-screenplay-etcd/internal/verdict"
)

// OfTheKey asks about the value a key holds:
//
//	err := anActor.Should(see.The(value.OfTheKey("app/config/database")).Should(be.EqualTo("postgresql://localhost:5432")))
//
// The question fails with an error wrapping etcd.ErrKeyNotFound when the key does
// not exist.
func OfTheKey(key string) screenplay.Question[string] {
	return &ofTheKeyQuestion{key: key}
}

// ofTheKeyQuestion asks about the value a key holds.
type ofTheKeyQuestion struct {
	key string
}

// AnsweredBy returns the value the key holds.
func (q *ofTheKeyQuestion) AnsweredBy(theActor *screenplay.Actor) (string, error) {
	key, err := internalquestion.RetrieveTheExistingKey(theActor, q.key)
	if err != nil {
		return "", err
	}

	return key.Value(), nil
}

// String describes the question.
func (q *ofTheKeyQuestion) String() string {
	return "value of the key " + verdict.Describe(q.key)
}

// ofTheKeyQuestion implements the screenplay.Question interface.
var _ screenplay.Question[string] = (*ofTheKeyQuestion)(nil)
