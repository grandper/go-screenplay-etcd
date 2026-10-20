package question

import (
	"github.com/grandper/go-screenplay/v2/screenplay"

	"github.com/grandper/go-screenplay-etcd/ability"
	impl "github.com/grandper/go-screenplay-etcd/internal/ability"
	"github.com/grandper/go-screenplay-etcd/internal/verdict"
)

// EventsOnTheKeysStartingWith asks about the changes recorded on the keys
// starting with a prefix the actor watches with watch.TheKeysStartingWith, in
// the order they happened:
//
//	err := anActor.Should(see.The(last.Of(question.EventsOnTheKeysStartingWith("app/config/"))).ShouldHave(been.Put()))
//
// The question fails with an error wrapping ability.ErrNotWatching when the
// actor does not watch the prefix.
func EventsOnTheKeysStartingWith(prefix string) screenplay.Question[[]*ability.Event] {
	return &eventsOnTheKeysStartingWithQuestion{prefix: prefix}
}

// eventsOnTheKeysStartingWithQuestion asks about the changes recorded on the
// watched keys starting with a prefix.
type eventsOnTheKeysStartingWithQuestion struct {
	prefix string
}

// AnsweredBy returns the events recorded on the keys starting with the prefix.
func (q *eventsOnTheKeysStartingWithQuestion) AnsweredBy(theActor *screenplay.Actor) ([]*ability.Event, error) {
	if q.prefix == "" {
		return nil, ErrPrefixNotProvided
	}

	return impl.TheAbilityOf(theActor).ToRetrieveTheEvents(q.prefix, true)
}

// String describes the question.
func (q *eventsOnTheKeysStartingWithQuestion) String() string {
	return "events on the keys starting with " + verdict.Describe(q.prefix)
}

// eventsOnTheKeysStartingWithQuestion implements the screenplay.Question interface.
var _ screenplay.Question[[]*ability.Event] = (*eventsOnTheKeysStartingWithQuestion)(nil)
