package ability

import (
	"context"

	"go.etcd.io/etcd/client/v3/concurrency"
)

// session returns the session the locks and the elections of the actor share,
// creating it on the first call. A session is a lease the client keeps alive:
// when it ends, every lock and every leadership of the actor is released.
func (o *Operations) session(ctx context.Context) (*concurrency.Session, error) {
	client, err := o.client(ctx)
	if err != nil {
		return nil, err
	}

	a := o.ability

	a.opening.Lock()
	defer a.opening.Unlock()

	if a.session == nil {
		session, sessionErr := concurrency.NewSession(client)
		if sessionErr != nil {
			return nil, sessionErr
		}

		a.session = session
	}

	return a.session, nil
}

// closeTheSession ends the session, releasing every lock and every leadership
// of the actor, and forgets them.
func (a *UseETCDAbility) closeTheSession() error {
	a.mutex.Lock()
	a.locks = nil
	a.elections = nil
	a.mutex.Unlock()

	a.opening.Lock()
	defer a.opening.Unlock()

	if a.session == nil {
		return nil
	}

	err := a.session.Close()
	a.session = nil

	return err
}
