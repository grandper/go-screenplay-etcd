package ability

import (
	"context"
	"errors"
	"fmt"

	"github.com/grandper/go-screenplay/v2/screenplay"
	"go.etcd.io/etcd/client/v3/concurrency"

	"github.com/grandper/go-screenplay-etcd/etcd"
	"github.com/grandper/go-screenplay-etcd/internal/verdict"
)

// ToCampaign campaigns in the election of the given name with the value, and
// returns once the actor leads the election, which may wait for the leader to
// resign. The leadership is given up by ToResign, or when the ability is
// forgotten. It fails with an error wrapping etcd.ErrAlreadyCampaigning when the
// actor already campaigns in the election.
func (o *Operations) ToCampaign(ctx context.Context, name, value string) error {
	session, err := o.session(ctx)
	if err != nil {
		return err
	}

	election, err := o.ability.addTheElection(session, name)
	if err != nil {
		return err
	}

	ctx, cancel := o.ability.context(ctx)
	defer cancel()

	if err = election.Campaign(ctx, value); err != nil {
		o.ability.removeTheElection(name)

		return err
	}

	return nil
}

// ToResign resigns from the election of the given name, giving up its
// leadership. It fails with an error wrapping etcd.ErrNotCampaigning when the actor
// does not campaign in the election.
func (o *Operations) ToResign(ctx context.Context, name string) error {
	if o.ability == nil {
		return screenplay.ErrActorMissingAbility
	}

	election, campaigning := o.ability.theElection(name)
	if !campaigning {
		return fmt.Errorf("%w in the election %s", etcd.ErrNotCampaigning, verdict.Describe(name))
	}

	ctx, cancel := o.ability.context(ctx)
	defer cancel()

	if err := election.Resign(ctx); err != nil {
		return err
	}

	o.ability.removeTheElection(name)

	return nil
}

// ToRetrieveTheLeader retrieves the value of the leader of the election of the
// given name, empty when the election has no leader.
func (o *Operations) ToRetrieveTheLeader(ctx context.Context, name string) (string, error) {
	session, err := o.session(ctx)
	if err != nil {
		return "", err
	}

	ctx, cancel := o.ability.context(ctx)
	defer cancel()

	response, err := concurrency.NewElection(session, name).Leader(ctx)
	if errors.Is(err, concurrency.ErrElectionNoLeader) {
		return "", nil
	}
	if err != nil {
		return "", err
	}

	if len(response.Kvs) == 0 {
		return "", nil
	}

	return string(response.Kvs[0].GetValue()), nil
}

// addTheElection registers the election the actor campaigns in, and fails
// with an error wrapping etcd.ErrAlreadyCampaigning when it is already registered.
func (a *UseETCDAbility) addTheElection(session *concurrency.Session, name string) (*concurrency.Election, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.elections == nil {
		a.elections = map[string]*concurrency.Election{}
	}

	if _, campaigning := a.elections[name]; campaigning {
		return nil, fmt.Errorf("%w in the election %s", etcd.ErrAlreadyCampaigning, verdict.Describe(name))
	}

	election := concurrency.NewElection(session, name)
	a.elections[name] = election

	return election, nil
}

// theElection returns the election the actor campaigns in, if any.
func (a *UseETCDAbility) theElection(name string) (*concurrency.Election, bool) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	election, campaigning := a.elections[name]

	return election, campaigning
}

// removeTheElection forgets the election.
func (a *UseETCDAbility) removeTheElection(name string) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	delete(a.elections, name)
}
