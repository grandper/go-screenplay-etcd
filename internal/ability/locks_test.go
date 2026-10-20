package ability_test

import (
	"testing"

	"github.com/grandper/go-screenplay/v2/screenplay"
	"github.com/stretchr/testify/require"

	"github.com/grandper/go-screenplay-etcd/fixture"
	"github.com/grandper/go-screenplay-etcd/internal/ability"
)

func TestLocks(t *testing.T) {
	server := fixture.RunETCDServer(t)
	newOperations := func(t *testing.T) *ability.Operations {
		t.Helper()
		theAbility := ability.UseETCD().WithTheEndpoints(server.Endpoint())
		t.Cleanup(func() { require.NoError(t, theAbility.Forget()) })

		return ability.TheAbilityOf(screenplay.ActorNamed("Alice").WhoCan(theAbility))
	}

	t.Run("acquires and releases a lock", func(t *testing.T) {
		alice, bob := newOperations(t), newOperations(t)

		require.NoError(t, alice.ToAcquireTheLock(t.Context(), "deployment", false))
		require.ErrorIs(t, bob.ToAcquireTheLock(t.Context(), "deployment", true), ability.ErrLocked)
		require.ErrorIs(t, alice.ToAcquireTheLock(t.Context(), "deployment", true), ability.ErrLockAlreadyHeld)

		require.NoError(t, alice.ToReleaseTheLock(t.Context(), "deployment"))
		require.ErrorIs(t, alice.ToReleaseTheLock(t.Context(), "deployment"), ability.ErrLockNotHeld)
		require.NoError(t, bob.ToAcquireTheLock(t.Context(), "deployment", false))
	})

	t.Run("releases the locks when the ability is forgotten", func(t *testing.T) {
		theAbility := ability.UseETCD().WithTheEndpoints(server.Endpoint())
		alice := ability.TheAbilityOf(screenplay.ActorNamed("Alice").WhoCan(theAbility))
		bob := newOperations(t)
		require.NoError(t, alice.ToAcquireTheLock(t.Context(), "release", false))

		require.NoError(t, theAbility.Forget())

		require.NoError(t, bob.ToAcquireTheLock(t.Context(), "release", false))
	})

	t.Run("fails every operation when the actor does not have the ability", func(t *testing.T) {
		noAbility := ability.TheAbilityOf(screenplay.ActorNamed("Bob"))

		require.ErrorIs(
			t,
			noAbility.ToAcquireTheLock(t.Context(), "deployment", true),
			screenplay.ErrActorMissingAbility,
		)
		require.ErrorIs(t, noAbility.ToReleaseTheLock(t.Context(), "deployment"), screenplay.ErrActorMissingAbility)
	})
}
