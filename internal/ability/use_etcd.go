package ability

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/grandper/go-screenplay/v2/action/timing"
	"github.com/grandper/go-screenplay/v2/screenplay"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/namespace"
)

// UseETCD enables the actor to use an etcd store. The ability is configured
// fluently, reading "use etcd with ...":
//
//	UseETCD().WithTheEndpoints("localhost:2379").WithTimingOutAfter(5).Seconds()
func UseETCD() *UseETCDAbility {
	return &UseETCDAbility{}
}

// UseETCDAbility is the ability to use an etcd store.
// It is safe for concurrent use: actions running in parallel (see
// proceed.ConcurrentlyTo) can put, retrieve and remove keys without racing.
// The client is created the first time the ability is used, so that a failure
// to connect is reported by the action that needed it. Every operation carries
// the actor's context, so that a deadline or a cancellation set on the actor
// stops it. Its method set holds only its settings and Forget; the operations
// are reached through TheAbilityOf, so a test cannot bypass the actor.
type UseETCDAbility struct {
	// connecting serializes the lazy connection, so that the state mutex is
	// never held during a network call.
	connecting sync.Mutex
	endpoints  []string
	namespace  string
	timeout    time.Duration
	client     *clientv3.Client
}

// WithTheEndpoints gives the ability the endpoints of the etcd cluster, in
// the form host:port; it reads "with the endpoints localhost:2379". The
// connection is opened the first time the ability is used.
func (a *UseETCDAbility) WithTheEndpoints(endpoints ...string) *UseETCDAbility {
	if a == nil {
		return nil
	}

	a.connecting.Lock()
	defer a.connecting.Unlock()

	a.endpoints = endpoints

	return a
}

// WithTheNamespace prefixes every key the actor uses with the namespace; it
// reads "with the namespace test-run-42/". The actions and the questions keep
// using plain keys: the prefix is added on the wire and removed from what the
// store answers, so that concurrent scenarios never share keys.
func (a *UseETCDAbility) WithTheNamespace(prefix string) *UseETCDAbility {
	if a == nil {
		return nil
	}

	a.connecting.Lock()
	defer a.connecting.Unlock()

	a.namespace = prefix

	return a
}

// WithTimingOutAfter bounds every operation of the ability; it reads "with
// timing out after 5 seconds". Without it, an operation is bounded only by
// the deadline of the actor's context.
func (a *UseETCDAbility) WithTimingOutAfter(amount int) *timing.DurationBuilder[UseETCDAbility] {
	if a == nil {
		return timing.NewDurationBuilder[UseETCDAbility](nil, nil).For(amount)
	}

	return timing.NewDurationBuilder(a, &a.timeout).For(amount)
}

// WithTheClient gives the ability a client the test created, instead of the
// endpoints; it reads "with the client". The namespace, if any, is set on the
// client. The ability owns the client from then on: it is closed when the
// actor exits.
func (a *UseETCDAbility) WithTheClient(client *clientv3.Client) *UseETCDAbility {
	if a == nil {
		return nil
	}

	a.connecting.Lock()
	defer a.connecting.Unlock()

	a.client = client

	return a
}

// Forget closes the connection to the store. On a nil receiver, the actor does
// not have the ability, it fails with screenplay.ErrActorMissingAbility.
func (a *UseETCDAbility) Forget() error {
	if a == nil {
		return screenplay.ErrActorMissingAbility
	}

	a.connecting.Lock()
	defer a.connecting.Unlock()

	if a.client == nil {
		return nil
	}

	err := a.client.Close()
	a.client = nil

	return err
}

// String describes the ability.
func (a *UseETCDAbility) String() string {
	return "use etcd"
}

// connect returns the client of the store, creating it from the endpoints on
// the first call. The namespace, if any, is set on the client once.
func (a *UseETCDAbility) connect() (*clientv3.Client, error) {
	a.connecting.Lock()
	defer a.connecting.Unlock()

	if a.client == nil {
		if len(a.endpoints) == 0 {
			return nil, ErrNoEndpointProvided
		}

		client, err := clientv3.New(clientv3.Config{Endpoints: a.endpoints})
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrFailedToConnect, err)
		}

		a.client = client
	}

	if a.namespace != "" {
		a.client.KV = namespace.NewKV(a.client.KV, a.namespace)
		a.client.Watcher = namespace.NewWatcher(a.client.Watcher, a.namespace)
		a.client.Lease = namespace.NewLease(a.client.Lease, a.namespace)
		a.namespace = ""
	}

	return a.client, nil
}

// context bounds the context of an operation with the timeout of the ability,
// when one was set.
func (a *UseETCDAbility) context(ctx context.Context) (context.Context, context.CancelFunc) {
	if a.timeout > 0 {
		return context.WithTimeout(ctx, a.timeout)
	}

	return context.WithCancel(ctx)
}

// UseETCDAbility implements the screenplay.Ability interface.
var _ screenplay.Ability = (*UseETCDAbility)(nil)

// TheAbilityOf returns the operations of the actor's ability to use etcd; it is
// written in PerformAs or AnsweredBy and reads "the ability of the actor to
// put":
//
//	err := ability.TheAbilityOf(theActor).ToPut(theActor.Context(), key, value)
//
// It never fails. When the actor is nil or does not have the ability, every
// operation fails with screenplay.ErrActorMissingAbility.
func TheAbilityOf(actor *screenplay.Actor) *Operations {
	return &Operations{ability: screenplay.UseAnAbilityOf(actor).That(UseETCD())}
}

// Operations holds the operations of the actor's ability to use etcd. They
// are not methods of UseETCDAbility, whose method set holds only its settings,
// so a test cannot reach the store without an action or a question.
type Operations struct {
	ability *UseETCDAbility
}

// client returns the client of the store, connecting on the first call. When
// the actor does not have the ability, it fails with
// screenplay.ErrActorMissingAbility.
func (o *Operations) client() (*clientv3.Client, error) {
	if o.ability == nil {
		return nil, screenplay.ErrActorMissingAbility
	}

	return o.ability.connect()
}

// ToPut puts the value for the key, creating the key or overwriting its value.
func (o *Operations) ToPut(ctx context.Context, key, value string) error {
	client, err := o.client()
	if err != nil {
		return err
	}

	ctx, cancel := o.ability.context(ctx)
	defer cancel()

	_, err = client.Put(ctx, key, value)

	return err
}

// ToRetrieveTheKey retrieves the key with its value and its metadata. A key
// that does not exist is answered as a Key that does not exist, not as an
// error.
func (o *Operations) ToRetrieveTheKey(ctx context.Context, key string) (*Key, error) {
	client, err := o.client()
	if err != nil {
		return nil, err
	}

	ctx, cancel := o.ability.context(ctx)
	defer cancel()

	response, err := client.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	if len(response.Kvs) == 0 {
		return NewMissingKey(key), nil
	}

	return NewKey(response.Kvs[0]), nil
}
