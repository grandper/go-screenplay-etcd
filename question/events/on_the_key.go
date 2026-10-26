package events

import (
	"github.com/grandper/go-screenplay/v2/screenplay"

	"github.com/grandper/go-screenplay-etcd/etcd"

	impl "github.com/grandper/go-screenplay-etcd/internal/ability"
	"github.com/grandper/go-screenplay-etcd/internal/verdict"
)

// OnTheKey asks about the changes recorded on a key the actor watches
// with watch.TheKey, in the order they happened:
//
//	err := anActor.Should(see.The(last.Of(events.OnTheKey("app/config/database"))).ShouldHave(been.Put()))
//
// The question fails with an error wrapping etcd.ErrNotWatching when the
// actor does not watch the key.
func OnTheKey(key string) screenplay.Question[[]*etcd.Event] {
	return &onTheKeyQuestion{key: key}
}

// onTheKeyQuestion asks about the changes recorded on a watched key.
type onTheKeyQuestion struct {
	key string
}

// AnsweredBy returns the events recorded on the key.
func (q *onTheKeyQuestion) AnsweredBy(theActor *screenplay.Actor) ([]*etcd.Event, error) {
	if q.key == "" {
		return nil, etcd.ErrKeyNotProvided
	}

	return impl.TheAbilityOf(theActor).ToRetrieveTheEvents(q.key, false)
}

// String describes the question.
func (q *onTheKeyQuestion) String() string {
	return "events on the key " + verdict.Describe(q.key)
}

// onTheKeyQuestion implements the screenplay.Question interface.
var _ screenplay.Question[[]*etcd.Event] = (*onTheKeyQuestion)(nil)
