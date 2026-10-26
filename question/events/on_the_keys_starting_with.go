package events

import (
	"github.com/grandper/go-screenplay/v2/screenplay"

	"github.com/grandper/go-screenplay-etcd/etcd"

	impl "github.com/grandper/go-screenplay-etcd/internal/ability"
	"github.com/grandper/go-screenplay-etcd/internal/verdict"
)

// OnTheKeysStartingWith asks about the changes recorded on the keys
// starting with a prefix the actor watches with watch.TheKeysStartingWith, in
// the order they happened:
//
//	err := anActor.Should(see.The(last.Of(events.OnTheKeysStartingWith("app/config/"))).ShouldHave(been.Put()))
//
// The question fails with an error wrapping etcd.ErrNotWatching when the
// actor does not watch the prefix.
func OnTheKeysStartingWith(prefix string) screenplay.Question[[]*etcd.Event] {
	return &onTheKeysStartingWithQuestion{prefix: prefix}
}

// onTheKeysStartingWithQuestion asks about the changes recorded on the
// watched keys starting with a prefix.
type onTheKeysStartingWithQuestion struct {
	prefix string
}

// AnsweredBy returns the events recorded on the keys starting with the prefix.
func (q *onTheKeysStartingWithQuestion) AnsweredBy(theActor *screenplay.Actor) ([]*etcd.Event, error) {
	if q.prefix == "" {
		return nil, etcd.ErrPrefixNotProvided
	}

	return impl.TheAbilityOf(theActor).ToRetrieveTheEvents(q.prefix, true)
}

// String describes the question.
func (q *onTheKeysStartingWithQuestion) String() string {
	return "events on the keys starting with " + verdict.Describe(q.prefix)
}

// onTheKeysStartingWithQuestion implements the screenplay.Question interface.
var _ screenplay.Question[[]*etcd.Event] = (*onTheKeysStartingWithQuestion)(nil)
