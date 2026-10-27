package ability_test

import (
	"testing"
	"time"

	"github.com/grandper/go-screenplay/v2/screenplay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grandper/go-screenplay-etcd/internal/ability"
	"github.com/grandper/go-screenplay-etcd/internal/fixture"
)

func TestLeases(t *testing.T) {
	server := fixture.RunETCDServer(t)

	t.Run("grants a lease and puts a key with it", func(t *testing.T) {
		theAbility := ability.TheAbilityOf(
			screenplay.ActorNamed("Alice").WhoCan(ability.UseETCD().WithTheEndpoints(server.Endpoint())),
		)

		lease, err := theAbility.ToGrantALease(t.Context(), 30*time.Second)
		require.NoError(t, err)
		assert.NotZero(t, lease)

		require.NoError(t, theAbility.ToPut(t.Context(), "locks/batch-job", "held", lease))

		key, err := theAbility.ToRetrieveTheKey(t.Context(), "locks/batch-job")
		require.NoError(t, err)
		assert.Equal(t, int64(lease), key.Lease())

		timeToLive, err := theAbility.ToRetrieveTheTimeToLive(t.Context(), lease)
		require.NoError(t, err)
		assert.Greater(t, timeToLive, 20*time.Second)
		assert.LessOrEqual(t, timeToLive, 30*time.Second)
	})

	t.Run("rounds a lease up to whole seconds", func(t *testing.T) {
		theAbility := ability.TheAbilityOf(
			screenplay.ActorNamed("Alice").WhoCan(ability.UseETCD().WithTheEndpoints(server.Endpoint())),
		)

		lease, err := theAbility.ToGrantALease(t.Context(), 2500*time.Millisecond)
		require.NoError(t, err)

		timeToLive, err := theAbility.ToRetrieveTheTimeToLive(t.Context(), lease)
		require.NoError(t, err)
		assert.LessOrEqual(t, timeToLive, 3*time.Second)
	})

	t.Run("answers zero for the time to live of a lease that expired", func(t *testing.T) {
		theAbility := ability.TheAbilityOf(
			screenplay.ActorNamed("Alice").WhoCan(ability.UseETCD().WithTheEndpoints(server.Endpoint())),
		)
		lease, err := theAbility.ToGrantALease(t.Context(), time.Second)
		require.NoError(t, err)

		assert.Eventually(t, func() bool {
			timeToLive, retrieveErr := theAbility.ToRetrieveTheTimeToLive(t.Context(), lease)

			return retrieveErr == nil && timeToLive == 0
		}, 15*time.Second, 200*time.Millisecond)
	})

	t.Run("keeps a lease alive until the ability is forgotten", func(t *testing.T) {
		theAbility := ability.UseETCD().WithTheEndpoints(server.Endpoint())
		operations := ability.TheAbilityOf(screenplay.ActorNamed("Alice").WhoCan(theAbility))
		lease, err := operations.ToGrantALease(t.Context(), time.Second)
		require.NoError(t, err)

		require.NoError(t, operations.ToKeepTheLeaseAlive(lease))

		time.Sleep(4 * time.Second)
		timeToLive, err := operations.ToRetrieveTheTimeToLive(t.Context(), lease)
		require.NoError(t, err)
		assert.Positive(t, timeToLive)

		require.NoError(t, theAbility.Forget())
		client := server.NewClient(t)
		assert.Eventually(t, func() bool {
			response, retrieveErr := client.TimeToLive(t.Context(), lease)

			return retrieveErr == nil && response.TTL <= 0
		}, 15*time.Second, 200*time.Millisecond)
	})

	t.Run("fails every operation when the actor does not have the ability", func(t *testing.T) {
		noAbility := ability.TheAbilityOf(screenplay.ActorNamed("Bob"))

		_, err := noAbility.ToGrantALease(t.Context(), time.Second)
		require.ErrorIs(t, err, screenplay.ErrActorMissingAbility)
		_, err = noAbility.ToRetrieveTheTimeToLive(t.Context(), 1)
		require.ErrorIs(t, err, screenplay.ErrActorMissingAbility)
		require.ErrorIs(t, noAbility.ToKeepTheLeaseAlive(1), screenplay.ErrActorMissingAbility)
	})
}
