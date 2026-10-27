package ability

import (
	"context"
	"fmt"
	"slices"
	"sync"

	"github.com/grandper/go-screenplay/v2/screenplay"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/grandper/go-screenplay-etcd/etcd"
	"github.com/grandper/go-screenplay-etcd/internal/answer"
	"github.com/grandper/go-screenplay-etcd/internal/verdict"
)

// watchID identifies a watch: a key, or the keys starting with a prefix.
type watchID struct {
	key        string
	withPrefix bool
}

// describe describes the watch, as "the key "k"" or "the keys starting with
// "p"", with the verb that follows it.
func (id watchID) describe() (string, string) {
	if id.withPrefix {
		return "the keys starting with " + verdict.Describe(id.key), "are"
	}

	return "the key " + verdict.Describe(id.key), "is"
}

// watch records the events notified on a key, or on the keys starting with a
// prefix, in the order they happened.
type watch struct {
	mutex  sync.RWMutex
	events []*answer.Event
}

// record appends an event to the watch.
func (w *watch) record(event *answer.Event) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.events = append(w.events, event)
}

// recorded returns a copy of the events of the watch, never nil.
func (w *watch) recorded() []*answer.Event {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	return slices.Clone(append([]*answer.Event{}, w.events...))
}

// ToWatch watches the key, or the keys starting with it when withPrefix is
// true, and records every event the store notifies until the ability is
// forgotten. It returns once the store confirmed the watch, so that a change
// made right after it is recorded. It fails with an error wrapping
// etcd.ErrAlreadyWatching when the key, or the prefix, is already watched.
func (o *Operations) ToWatch(ctx context.Context, key string, withPrefix bool) error {
	client, err := o.client()
	if err != nil {
		return err
	}

	id := watchID{key: key, withPrefix: withPrefix}

	theWatch, err := o.ability.addTheWatch(id)
	if err != nil {
		return err
	}

	created := make(chan struct{})
	ended := make(chan struct{})

	o.ability.background.run(func(ctx context.Context) {
		defer close(ended)

		o.recordTheEvents(ctx, client, id, theWatch, created)
	})

	ctx, cancel := o.ability.context(ctx)
	defer cancel()

	select {
	case <-created:
		return nil
	case <-ended:
		return fmt.Errorf("%w: the watch was canceled", etcd.ErrFailedToConnect)
	case <-ctx.Done():
		return ctx.Err()
	}
}

// recordTheEvents watches with the client and records every event notified
// into the watch, until the context is done. It closes created once the store
// confirmed the watch.
func (o *Operations) recordTheEvents(
	ctx context.Context,
	client *clientv3.Client,
	id watchID,
	theWatch *watch,
	created chan<- struct{},
) {
	options := []clientv3.OpOption{clientv3.WithPrevKV(), clientv3.WithCreatedNotify()}
	if id.withPrefix {
		options = append(options, clientv3.WithPrefix())
	}

	confirmed := false
	for response := range client.Watch(ctx, id.key, options...) {
		if response.Created && !confirmed {
			confirmed = true
			close(created)
		}

		for _, event := range response.Events {
			theWatch.record(answer.NewEvent(event, o.causedByLeaseExpiry(ctx, client, event)))
		}
	}
}

// causedByLeaseExpiry tells whether a delete event was caused by the expiry
// (or the revocation) of the lease of the key: the key had a lease, and the
// lease is gone.
func (o *Operations) causedByLeaseExpiry(ctx context.Context, client *clientv3.Client, event *clientv3.Event) bool {
	if event.GetType() != mvccpb.DELETE || event.GetPrevKv().GetLease() == 0 {
		return false
	}

	ctx, cancel := o.ability.context(ctx)
	defer cancel()

	response, err := client.TimeToLive(ctx, clientv3.LeaseID(event.GetPrevKv().GetLease()))
	if err != nil {
		return false
	}

	return response.TTL < 0
}

// ToRetrieveTheEvents returns the events recorded so far on the watched key,
// or on the watched keys starting with it when withPrefix is true, in the
// order they happened. It fails with an error wrapping etcd.ErrNotWatching when the
// key, or the prefix, is not watched.
func (o *Operations) ToRetrieveTheEvents(key string, withPrefix bool) ([]*answer.Event, error) {
	if o.ability == nil {
		return nil, screenplay.ErrActorMissingAbility
	}

	id := watchID{key: key, withPrefix: withPrefix}

	theWatch, watched := o.ability.theWatch(id)
	if !watched {
		subject, verb := id.describe()

		return nil, fmt.Errorf("%s %s %w", subject, verb, etcd.ErrNotWatching)
	}

	return theWatch.recorded(), nil
}

// addTheWatch registers a watch on the ability, and fails with an error
// wrapping etcd.ErrAlreadyWatching when one is already registered.
func (a *UseETCDAbility) addTheWatch(id watchID) (*watch, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.watches == nil {
		a.watches = map[watchID]*watch{}
	}

	if _, watched := a.watches[id]; watched {
		subject, verb := id.describe()

		return nil, fmt.Errorf("%s %s %w", subject, verb, etcd.ErrAlreadyWatching)
	}

	theWatch := &watch{}
	a.watches[id] = theWatch

	return theWatch, nil
}

// theWatch returns the watch registered on the ability, if any.
func (a *UseETCDAbility) theWatch(id watchID) (*watch, bool) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	theWatch, watched := a.watches[id]

	return theWatch, watched
}

// forgetTheWatches drops every watch, once their goroutines are stopped.
func (a *UseETCDAbility) forgetTheWatches() {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.watches = nil
}
