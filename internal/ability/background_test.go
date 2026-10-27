package ability_test

import (
	"testing"
	"time"

	"github.com/grandper/go-screenplay/v2/screenplay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/grandper/go-screenplay-etcd/etcd"
	"github.com/grandper/go-screenplay-etcd/fixture"
	"github.com/grandper/go-screenplay-etcd/internal/ability"
)

func TestBackground(t *testing.T) {
	server := fixture.RunETCDServer(t)
	client := server.NewClient(t)
	// isExpired tells whether the lease is gone from the store, which answers
	// with a negative time to live; a live lease answers with zero between two
	// renewals when it lives for one second.
	isExpired := func(t *testing.T, lease clientv3.LeaseID) bool {
		t.Helper()
		response, err := client.TimeToLive(t.Context(), lease)

		return err == nil && response.TTL < 0
	}

	t.Run("stops every goroutine of the ability when it is forgotten", func(t *testing.T) {
		theAbility := ability.UseETCD().WithTheEndpoints(server.Endpoint())
		alice := ability.TheAbilityOf(screenplay.ActorNamed("Alice").WhoCan(theAbility))
		first, err := alice.ToGrantALease(t.Context(), time.Second)
		require.NoError(t, err)
		require.NoError(t, alice.ToKeepTheLeaseAlive(first))
		second, err := alice.ToGrantALease(t.Context(), time.Second)
		require.NoError(t, err)
		require.NoError(t, alice.ToKeepTheLeaseAlive(second))
		require.NoError(t, alice.ToWatch(t.Context(), "app/status", false))

		time.Sleep(3 * time.Second)
		assert.False(t, isExpired(t, first), "the first lease is renewed in the background")
		assert.False(t, isExpired(t, second), "the second lease is renewed in the background")

		require.NoError(t, theAbility.Forget())

		assert.Eventually(t, func() bool {
			return isExpired(t, first) && isExpired(t, second)
		}, 15*time.Second, 200*time.Millisecond, "the renewals should stop once the ability is forgotten")
		_, err = alice.ToRetrieveTheEvents("app/status", false)
		require.ErrorIs(t, err, etcd.ErrNotWatching, "the watch should stop once the ability is forgotten")
	})

	t.Run("runs nothing when the actor never used it", func(t *testing.T) {
		theAbility := ability.UseETCD().WithTheEndpoints(server.Endpoint())

		require.NoError(t, theAbility.Forget())
	})

	t.Run("is stopped once, however many times the ability is forgotten", func(t *testing.T) {
		theAbility := ability.UseETCD().WithTheEndpoints(server.Endpoint())
		alice := ability.TheAbilityOf(screenplay.ActorNamed("Alice").WhoCan(theAbility))
		lease, err := alice.ToGrantALease(t.Context(), time.Second)
		require.NoError(t, err)
		require.NoError(t, alice.ToKeepTheLeaseAlive(lease))

		require.NoError(t, theAbility.Forget())
		require.NoError(t, theAbility.Forget())

		assert.Eventually(t, func() bool { return isExpired(t, lease) }, 15*time.Second, 200*time.Millisecond)
	})

	t.Run("starts new goroutines after the ability was forgotten", func(t *testing.T) {
		theAbility := ability.UseETCD().WithTheEndpoints(server.Endpoint())
		alice := ability.TheAbilityOf(screenplay.ActorNamed("Alice").WhoCan(theAbility))
		require.NoError(t, alice.ToPut(t.Context(), "app/status", "ready", 0))
		require.NoError(t, theAbility.Forget())
		t.Cleanup(func() { require.NoError(t, theAbility.Forget()) })

		lease, err := alice.ToGrantALease(t.Context(), time.Second)
		require.NoError(t, err)
		require.NoError(t, alice.ToKeepTheLeaseAlive(lease))

		time.Sleep(3 * time.Second)
		assert.False(t, isExpired(t, lease), "the lease is renewed in the background")
	})
}
