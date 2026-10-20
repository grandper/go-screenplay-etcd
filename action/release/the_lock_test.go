package release_test

import (
	"testing"

	"github.com/grandper/go-screenplay/v2/screenplay"
	screenplayfixture "github.com/grandper/go-screenplay/v2/screenplay/fixture"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grandper/go-screenplay-etcd/ability"
	"github.com/grandper/go-screenplay-etcd/action"
	"github.com/grandper/go-screenplay-etcd/action/acquire"
	"github.com/grandper/go-screenplay-etcd/action/release"
	"github.com/grandper/go-screenplay-etcd/fixture"
	etcdfixture "github.com/grandper/go-screenplay-etcd/internal/fixture"
)

func TestTheLock(t *testing.T) {
	server := fixture.RunETCDServer(t)
	newActor := func(t *testing.T, name string) *screenplay.Actor {
		t.Helper()
		actor := screenplay.ActorNamed(name).WhoCan(ability.UseETCD().WithTheEndpoints(server.Endpoint()))
		t.Cleanup(func() { require.NoError(t, actor.Exit()) })

		return actor
	}

	t.Run("releases a lock", func(t *testing.T) {
		t.Run("the actor holds, so that another actor can acquire it", func(t *testing.T) {
			alice, bob := newActor(t, "Alice"), newActor(t, "Bob")
			require.NoError(t, alice.AttemptsTo(acquire.TheLock("deployment")))

			require.NoError(t, alice.AttemptsTo(release.TheLock("deployment")))

			require.NoError(t, bob.AttemptsTo(acquire.TheLock("deployment").WithoutWaiting()))
		})

		t.Run("that can be acquired again by the actor", func(t *testing.T) {
			alice := newActor(t, "Alice")

			require.NoError(t, alice.AttemptsTo(
				acquire.TheLock("deployment"),
				release.TheLock("deployment"),
				acquire.TheLock("deployment"),
			))
		})

		t.Run("fails when the actor does not hold it", func(t *testing.T) {
			alice := newActor(t, "Alice")

			err := alice.AttemptsTo(release.TheLock("deployment"))

			require.ErrorIs(t, err, ability.ErrLockNotHeld)
			require.EqualError(t, err, `the lock "deployment" is not held by the actor`)
		})

		t.Run("fails when the name is not provided", func(t *testing.T) {
			alice := newActor(t, "Alice")

			require.ErrorIs(t, alice.AttemptsTo(release.TheLock("")), action.ErrNameNotProvided)
		})

		t.Run("fails when the actor cannot use etcd", func(t *testing.T) {
			err := screenplay.ActorNamed("Carol").AttemptsTo(release.TheLock("deployment"))

			require.ErrorIs(t, err, screenplay.ErrActorMissingAbility)
		})
	})

	t.Run("is narrated", func(t *testing.T) {
		recorder := screenplayfixture.NewRecorder()
		adam := etcdfixture.NarratedActor("Adam", recorder, server.Endpoint())
		t.Cleanup(func() { require.NoError(t, adam.Exit()) })

		require.NoError(t, adam.AttemptsTo(acquire.TheLock("deployment"), release.TheLock("deployment")))

		assert.Equal(t, `Adam attempts to release the lock "deployment"`, recorder.Messages()[2])
	})

	t.Run("implements the stringer interface", func(t *testing.T) {
		assert.Equal(t, `release the lock "deployment"`, release.TheLock("deployment").String())
	})
}
