package acquire_test

import (
	"testing"

	"github.com/grandper/go-screenplay/v2/screenplay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/grandper/go-screenplay-etcd/ability/use"
	"github.com/grandper/go-screenplay-etcd/etcd"
	"github.com/grandper/go-screenplay-etcd/internal/action/acquire"
	"github.com/grandper/go-screenplay-etcd/internal/fixture"
)

func TestTheLock(t *testing.T) {
	server := fixture.RunETCDServer(t)
	client := server.NewClient(t)
	newActor := func(t *testing.T, name string) *screenplay.Actor {
		t.Helper()
		actor := screenplay.ActorNamed(name).WhoCan(use.ETCD().WithTheEndpoints(server.Endpoint()))
		t.Cleanup(func() { require.NoError(t, actor.Exit()) })

		return actor
	}

	t.Run("acquires the lock, holding it under its name in the store", func(t *testing.T) {
		alice := newActor(t, "Alice")

		require.NoError(t, alice.AttemptsTo(acquire.TheLock("deployment")))

		response, err := client.Get(t.Context(), "deployment/", clientv3.WithPrefix())
		require.NoError(t, err)
		assert.Len(t, response.Kvs, 1, "the lock is a key under its name")
	})

	t.Run("fails without waiting when another actor holds the lock", func(t *testing.T) {
		alice, bob := newActor(t, "Alice"), newActor(t, "Bob")
		require.NoError(t, alice.AttemptsTo(acquire.TheLock("deployment")))

		err := bob.AttemptsTo(acquire.TheLock("deployment").WithoutWaiting())

		require.ErrorIs(t, err, etcd.ErrLocked)
	})

	t.Run("fails when the actor already holds the lock", func(t *testing.T) {
		alice := newActor(t, "Alice")
		require.NoError(t, alice.AttemptsTo(acquire.TheLock("deployment")))

		err := alice.AttemptsTo(acquire.TheLock("deployment").WithoutWaiting())

		require.ErrorIs(t, err, etcd.ErrLockAlreadyHeld)
		require.EqualError(t, err, `the lock "deployment" is already held by the actor`)
	})

	t.Run("reports a missing name before reaching the store", func(t *testing.T) {
		err := screenplay.ActorNamed("Carol").AttemptsTo(acquire.TheLock(""))

		require.ErrorIs(t, err, etcd.ErrNameNotProvided)
		require.NotErrorIs(t, err, screenplay.ErrActorMissingAbility)
	})

	t.Run("fails when the actor cannot use etcd", func(t *testing.T) {
		err := screenplay.ActorNamed("Carol").AttemptsTo(acquire.TheLock("deployment"))

		require.ErrorIs(t, err, screenplay.ErrActorMissingAbility)
	})

	t.Run("implements the stringer interface", func(t *testing.T) {
		assert.Equal(t, `acquire the lock "deployment"`, acquire.TheLock("deployment").String())
		assert.Equal(
			t,
			`acquire the lock "deployment" without waiting`,
			acquire.TheLock("deployment").WithoutWaiting().String(),
		)
	})

	t.Run("implements the performable interface", func(t *testing.T) {
		assert.Implements(t, (*screenplay.Performable)(nil), acquire.TheLock("deployment"))
	})
}
