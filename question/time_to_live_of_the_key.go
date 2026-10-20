package question

import (
	"time"

	"github.com/grandper/go-screenplay/v2/screenplay"
	clientv3 "go.etcd.io/etcd/client/v3"

	impl "github.com/grandper/go-screenplay-etcd/internal/ability"
	"github.com/grandper/go-screenplay-etcd/internal/verdict"
)

// TimeToLiveOfTheKey asks about the time a key has left to live before its
// lease expires and removes it, in whole seconds. A key that never expires,
// having no lease, has zero time to live:
//
//	err := anActor.Should(see.The(question.TimeToLiveOfTheKey("locks/batch-job")).Should(be.AtMost(10*time.Second)))
//
// The question fails with an error wrapping ErrKeyNotFound when the key does
// not exist.
func TimeToLiveOfTheKey(key string) screenplay.Question[time.Duration] {
	return &timeToLiveOfTheKeyQuestion{key: key}
}

// timeToLiveOfTheKeyQuestion asks about the time a key has left to live.
type timeToLiveOfTheKeyQuestion struct {
	key string
}

// AnsweredBy returns the time the key has left to live.
func (q *timeToLiveOfTheKeyQuestion) AnsweredBy(theActor *screenplay.Actor) (time.Duration, error) {
	key, err := retrieveTheExistingKey(theActor, q.key)
	if err != nil {
		return 0, err
	}

	if key.Lease() == 0 {
		return 0, nil
	}

	return impl.TheAbilityOf(theActor).ToRetrieveTheTimeToLive(theActor.Context(), clientv3.LeaseID(key.Lease()))
}

// String describes the question.
func (q *timeToLiveOfTheKeyQuestion) String() string {
	return "time to live of the key " + verdict.Describe(q.key)
}

// timeToLiveOfTheKeyQuestion implements the screenplay.Question interface.
var _ screenplay.Question[time.Duration] = (*timeToLiveOfTheKeyQuestion)(nil)
