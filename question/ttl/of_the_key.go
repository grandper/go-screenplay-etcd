package ttl

import (
	"time"

	internalquestion "github.com/grandper/go-screenplay-etcd/internal/question"

	"github.com/grandper/go-screenplay/v2/screenplay"
	clientv3 "go.etcd.io/etcd/client/v3"

	impl "github.com/grandper/go-screenplay-etcd/internal/ability"
	"github.com/grandper/go-screenplay-etcd/internal/verdict"
)

// OfTheKey asks about the time a key has left to live before its
// lease expires and removes it, in whole seconds. A key that never expires,
// having no lease, has zero time to live:
//
//	err := anActor.Should(see.The(ttl.OfTheKey("locks/batch-job")).Should(be.AtMost(10*time.Second)))
//
// The question fails with an error wrapping question.ErrKeyNotFound when the key does
// not exist.
func OfTheKey(key string) screenplay.Question[time.Duration] {
	return &ofTheKeyQuestion{key: key}
}

// ofTheKeyQuestion asks about the time a key has left to live.
type ofTheKeyQuestion struct {
	key string
}

// AnsweredBy returns the time the key has left to live.
func (q *ofTheKeyQuestion) AnsweredBy(theActor *screenplay.Actor) (time.Duration, error) {
	key, err := internalquestion.RetrieveTheExistingKey(theActor, q.key)
	if err != nil {
		return 0, err
	}

	if key.Lease() == 0 {
		return 0, nil
	}

	return impl.TheAbilityOf(theActor).ToRetrieveTheTimeToLive(theActor.Context(), clientv3.LeaseID(key.Lease()))
}

// String describes the question.
func (q *ofTheKeyQuestion) String() string {
	return "TTL of the key " + verdict.Describe(q.key)
}

// ofTheKeyQuestion implements the screenplay.Question interface.
var _ screenplay.Question[time.Duration] = (*ofTheKeyQuestion)(nil)
