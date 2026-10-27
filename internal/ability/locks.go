package ability

import (
	"context"
	"fmt"

	"go.etcd.io/etcd/client/v3/concurrency"

	"github.com/grandper/go-screenplay/v2/screenplay"

	"github.com/grandper/go-screenplay-etcd/etcd"
	"github.com/grandper/go-screenplay-etcd/internal/verdict"
)

// ToAcquireTheLock acquires the distributed lock of the given name, waiting
// for the actor holding it to release it unless withoutWaiting is true, in
// which case it fails with etcd.ErrLocked. The lock is released by
// ToReleaseTheLock, or when the ability is forgotten. It fails with an error
// wrapping etcd.ErrLockAlreadyHeld when the actor already holds the lock.
func (o *Operations) ToAcquireTheLock(ctx context.Context, name string, withoutWaiting bool) error {
	session, err := o.session()
	if err != nil {
		return err
	}

	lock, err := o.ability.addTheLock(session, name)
	if err != nil {
		return err
	}

	ctx, cancel := o.ability.context(ctx)
	defer cancel()

	if withoutWaiting {
		err = lock.TryLock(ctx)
	} else {
		err = lock.Lock(ctx)
	}

	if err != nil {
		o.ability.removeTheLock(name)

		return err
	}

	return nil
}

// ToReleaseTheLock releases the distributed lock of the given name. It fails
// with an error wrapping etcd.ErrLockNotHeld when the actor does not hold the lock.
func (o *Operations) ToReleaseTheLock(ctx context.Context, name string) error {
	if o.ability == nil {
		return screenplay.ErrActorMissingAbility
	}

	lock, held := o.ability.theLock(name)
	if !held {
		return fmt.Errorf("the lock %s is %w", verdict.Describe(name), etcd.ErrLockNotHeld)
	}

	ctx, cancel := o.ability.context(ctx)
	defer cancel()

	if err := lock.Unlock(ctx); err != nil {
		return err
	}

	o.ability.removeTheLock(name)

	return nil
}

// addTheLock registers the lock the actor is acquiring, and fails with an
// error wrapping etcd.ErrLockAlreadyHeld when it is already registered.
func (a *UseETCDAbility) addTheLock(session *concurrency.Session, name string) (*concurrency.Mutex, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.locks == nil {
		a.locks = map[string]*concurrency.Mutex{}
	}

	if _, held := a.locks[name]; held {
		return nil, fmt.Errorf("the lock %s is %w", verdict.Describe(name), etcd.ErrLockAlreadyHeld)
	}

	lock := concurrency.NewMutex(session, name)
	a.locks[name] = lock

	return lock, nil
}

// theLock returns the lock the actor holds, if any.
func (a *UseETCDAbility) theLock(name string) (*concurrency.Mutex, bool) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	lock, held := a.locks[name]

	return lock, held
}

// removeTheLock forgets the lock.
func (a *UseETCDAbility) removeTheLock(name string) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	delete(a.locks, name)
}
