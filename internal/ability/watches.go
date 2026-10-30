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
	// err is set once the store canceled the watch: the events that followed
	// were not recorded, so the watch can no longer answer for them.
	err error
}

// record appends an event to the watch.
func (w *watch) record(event *answer.Event) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.events = append(w.events, event)
}

// end marks the watch as canceled by the store, with the error that says so.
func (w *watch) end(err error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.err = err
}

// recorded returns a copy of the events of the watch, never nil, or the error
// of the watch once the store canceled it.
func (w *watch) recorded() ([]*answer.Event, error) {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	if w.err != nil {
		return nil, w.err
	}

	return slices.Clone(append([]*answer.Event{}, w.events...)), nil
}

// ToWatch watches the key, or the keys starting with it when withPrefix is
// true, and records every event the store notifies until the ability is
// forgotten. It returns once the store confirmed the watch, so that a change
// made right after it is recorded; when the context ends first, the watch is
// dropped, so that it can be attempted again. It fails with an error wrapping
// etcd.ErrAlreadyWatching when the key, or the prefix, is already watched.
func (o *Operations) ToWatch(ctx context.Context, key string, withPrefix bool) error {
	client, err := o.client(ctx)
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

	stop := o.ability.background.run(func(ctx context.Context) {
		defer close(ended)

		o.recordTheEvents(ctx, client, id, theWatch, created)
	})

	ctx, cancel := o.ability.context(ctx)
	defer cancel()

	select {
	case <-created:
		return nil
	case <-ended:
		o.ability.removeTheWatch(id)

		if _, err = theWatch.recorded(); err != nil {
			return err
		}

		return canceledWatchError(id, nil)
	case <-ctx.Done():
		stop()
		<-ended
		o.ability.removeTheWatch(id)

		return ctx.Err()
	}
}

// recordTheEvents watches with the client and records every event notified
// into the watch, until the context is done. It closes created once the store
// confirmed the watch. When the store cancels the watch instead, the watch is
// ended with the cause, as the events that follow are not recorded.
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

	var cause error

	for response := range client.Watch(ctx, id.key, options...) {
		if response.Created && !confirmed {
			confirmed = true
			close(created)
		}

		if err := response.Err(); err != nil {
			cause = err
		}

		for _, event := range response.Events {
			theWatch.record(answer.NewEvent(event, o.causedByLeaseExpiry(ctx, client, event)))
		}
	}

	if ctx.Err() != nil {
		// The ability stopped the watch itself.
		return
	}

	theWatch.end(canceledWatchError(id, cause))
}

// canceledWatchError describes a watch the store canceled, with the cause the
// store gave, if any. It wraps etcd.ErrFailedToConnect.
func canceledWatchError(id watchID, cause error) error {
	subject, _ := id.describe()

	if cause == nil {
		return fmt.Errorf("%w: the watch of %s was canceled", etcd.ErrFailedToConnect, subject)
	}

	return fmt.Errorf("%w: the watch of %s was canceled: %w", etcd.ErrFailedToConnect, subject, cause)
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
// order they happened. It fails with an error wrapping etcd.ErrNotWatching
// when the key, or the prefix, is not watched, and with an error wrapping
// etcd.ErrFailedToConnect once the store canceled the watch, as the events
// that followed were not recorded.
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

	return theWatch.recorded()
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

// removeTheWatch drops the watch, once its goroutine is stopped.
func (a *UseETCDAbility) removeTheWatch(id watchID) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	delete(a.watches, id)
}

// forgetTheWatches drops every watch, once their goroutines are stopped.
func (a *UseETCDAbility) forgetTheWatches() {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.watches = nil
}
