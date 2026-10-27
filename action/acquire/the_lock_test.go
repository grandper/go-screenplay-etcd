package acquire_test

import (
	"context"
	"testing"
	"time"

	"github.com/grandper/go-screenplay/v2/screenplay"
	screenplayfixture "github.com/grandper/go-screenplay/v2/screenplay/fixture"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grandper/go-screenplay-etcd/ability/use"
	"github.com/grandper/go-screenplay-etcd/action/acquire"
	"github.com/grandper/go-screenplay-etcd/action/release"
	"github.com/grandper/go-screenplay-etcd/etcd"
	"github.com/grandper/go-screenplay-etcd/internal/fixture"
)

func TestTheLock(t *testing.T) {
	server := fixture.RunETCDServer(t)
	newActor := func(t *testing.T, name string) *screenplay.Actor {
		t.Helper()
		actor := screenplay.ActorNamed(name).WhoCan(use.ETCD().WithTheEndpoints(server.Endpoint()))
		t.Cleanup(func() { require.NoError(t, actor.Exit()) })

		return actor
	}

	t.Run("acquires a lock", func(t *testing.T) {
		t.Run("that nobody holds", func(t *testing.T) {
			alice := newActor(t, "Alice")

			require.NoError(t, alice.AttemptsTo(acquire.TheLock("deployment")))
		})

		t.Run("waiting for the actor holding it to release it", func(t *testing.T) {
			alice, bob := newActor(t, "Alice"), newActor(t, "Bob")
			require.NoError(t, alice.AttemptsTo(acquire.TheLock("deployment")))

			acquired := make(chan error, 1)
			go func() { acquired <- bob.AttemptsTo(acquire.TheLock("deployment")) }()

			select {
			case err := <-acquired:
				t.Fatalf("Bob acquired the lock Alice holds: %v", err)
			case <-time.After(500 * time.Millisecond):
			}

			require.NoError(t, alice.AttemptsTo(release.TheLock("deployment")))
			select {
			case err := <-acquired:
				require.NoError(t, err)
			case <-time.After(15 * time.Second):
				t.Fatal("Bob did not acquire the lock Alice released")
			}
		})

		t.Run("without waiting, failing when another actor holds it", func(t *testing.T) {
			alice, bob := newActor(t, "Alice"), newActor(t, "Bob")
			require.NoError(t, alice.AttemptsTo(acquire.TheLock("deployment")))

			err := bob.AttemptsTo(acquire.TheLock("deployment").WithoutWaiting())

			require.ErrorIs(t, err, etcd.ErrLocked)
		})

		t.Run("waiting no longer than the context of the actor allows", func(t *testing.T) {
			alice, bob := newActor(t, "Alice"), newActor(t, "Bob")
			require.NoError(t, alice.AttemptsTo(acquire.TheLock("deployment")))
			ctx, cancel := context.WithTimeout(t.Context(), 500*time.Millisecond)
			defer cancel()

			err := bob.WithContext(ctx).AttemptsTo(acquire.TheLock("deployment"))

			require.ErrorIs(t, err, context.DeadlineExceeded)
		})

		t.Run("released when the actor exits", func(t *testing.T) {
			alice := screenplay.ActorNamed("Alice").WhoCan(use.ETCD().WithTheEndpoints(server.Endpoint()))
			bob := newActor(t, "Bob")
			require.NoError(t, alice.AttemptsTo(acquire.TheLock("deployment")))

			require.NoError(t, alice.Exit())

			require.NoError(t, bob.AttemptsTo(acquire.TheLock("deployment").WithoutWaiting()))
		})

		t.Run("fails when the actor already holds it", func(t *testing.T) {
			alice := newActor(t, "Alice")
			require.NoError(t, alice.AttemptsTo(acquire.TheLock("deployment")))

			err := alice.AttemptsTo(acquire.TheLock("deployment"))

			require.ErrorIs(t, err, etcd.ErrLockAlreadyHeld)
			require.EqualError(t, err, `the lock "deployment" is already held by the actor`)
		})

		t.Run("fails when the name is not provided", func(t *testing.T) {
			alice := newActor(t, "Alice")

			require.ErrorIs(t, alice.AttemptsTo(acquire.TheLock("")), etcd.ErrNameNotProvided)
		})

		t.Run("fails when the actor cannot use etcd", func(t *testing.T) {
			err := screenplay.ActorNamed("Carol").AttemptsTo(acquire.TheLock("deployment"))

			require.ErrorIs(t, err, screenplay.ErrActorMissingAbility)
		})
	})

	t.Run("is narrated", func(t *testing.T) {
		recorder := screenplayfixture.NewRecorder()
		adam := fixture.NarratedActor("Adam", recorder, server.Endpoint())
		t.Cleanup(func() { require.NoError(t, adam.Exit()) })

		require.NoError(t, adam.AttemptsTo(acquire.TheLock("deployment")))

		assert.Equal(t, `Adam attempts to acquire the lock "deployment"`, recorder.Messages()[0])
	})

	t.Run("implements the stringer interface", func(t *testing.T) {
		assert.Equal(t, `acquire the lock "deployment"`, acquire.TheLock("deployment").String())
		assert.Equal(
			t,
			`acquire the lock "deployment" without waiting`,
			acquire.TheLock("deployment").WithoutWaiting().String(),
		)
	})
}
