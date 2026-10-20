package events

import (
	"github.com/grandper/go-screenplay/v2/screenplay"

	"github.com/grandper/go-screenplay-etcd/question"

	"github.com/grandper/go-screenplay-etcd/ability"
	impl "github.com/grandper/go-screenplay-etcd/internal/ability"
	"github.com/grandper/go-screenplay-etcd/internal/verdict"
)

// OnTheKey asks about the changes recorded on a key the actor watches
// with watch.TheKey, in the order they happened:
//
//	err := anActor.Should(see.The(last.Of(events.OnTheKey("app/config/database"))).ShouldHave(been.Put()))
//
// The question fails with an error wrapping ability.ErrNotWatching when the
// actor does not watch the key.
func OnTheKey(key string) screenplay.Question[[]*ability.Event] {
	return &onTheKeyQuestion{key: key}
}

// onTheKeyQuestion asks about the changes recorded on a watched key.
type onTheKeyQuestion struct {
	key string
}

// AnsweredBy returns the events recorded on the key.
func (q *onTheKeyQuestion) AnsweredBy(theActor *screenplay.Actor) ([]*ability.Event, error) {
	if q.key == "" {
		return nil, question.ErrKeyNotProvided
	}

	return impl.TheAbilityOf(theActor).ToRetrieveTheEvents(q.key, false)
}

// String describes the question.
func (q *onTheKeyQuestion) String() string {
	return "events on the key " + verdict.Describe(q.key)
}

// onTheKeyQuestion implements the screenplay.Question interface.
var _ screenplay.Question[[]*ability.Event] = (*onTheKeyQuestion)(nil)
