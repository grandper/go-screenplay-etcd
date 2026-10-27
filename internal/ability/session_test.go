package ability_test

import (
	"testing"

	"github.com/grandper/go-screenplay/v2/screenplay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/grandper/go-screenplay-etcd/etcd"
	"github.com/grandper/go-screenplay-etcd/fixture"
	"github.com/grandper/go-screenplay-etcd/internal/ability"
)

func TestSession(t *testing.T) {
	server := fixture.RunETCDServer(t)
	client := server.NewClient(t)
	// leaseOfTheOnlyKeyUnder returns the lease of the key the lock, or the
	// election, of the given name is held with.
	leaseOfTheOnlyKeyUnder := func(t *testing.T, name string) int64 {
		t.Helper()
		response, err := client.Get(t.Context(), name+"/", clientv3.WithPrefix())
		require.NoError(t, err)
		require.Len(t, response.Kvs, 1)

		return response.Kvs[0].GetLease()
	}

	t.Run("is shared by the locks and the elections of the actor", func(t *testing.T) {
		theAbility := ability.UseETCD().WithTheEndpoints(server.Endpoint())
		t.Cleanup(func() { require.NoError(t, theAbility.Forget()) })
		alice := ability.TheAbilityOf(screenplay.ActorNamed("Alice").WhoCan(theAbility))

		require.NoError(t, alice.ToAcquireTheLock(t.Context(), "deployment", false))
		require.NoError(t, alice.ToCampaign(t.Context(), "primary-worker", "Alice"))

		lease := leaseOfTheOnlyKeyUnder(t, "deployment")
		assert.NotZero(t, lease)
		assert.Equal(t, lease, leaseOfTheOnlyKeyUnder(t, "primary-worker"))
	})

	t.Run("is not shared between actors", func(t *testing.T) {
		alice := ability.UseETCD().WithTheEndpoints(server.Endpoint())
		t.Cleanup(func() { require.NoError(t, alice.Forget()) })
		bob := ability.UseETCD().WithTheEndpoints(server.Endpoint())
		t.Cleanup(func() { require.NoError(t, bob.Forget()) })

		require.NoError(
			t,
			ability.TheAbilityOf(screenplay.ActorNamed("Alice").WhoCan(alice)).
				ToAcquireTheLock(t.Context(), "a", false),
		)
		require.NoError(
			t,
			ability.TheAbilityOf(screenplay.ActorNamed("Bob").WhoCan(bob)).ToAcquireTheLock(t.Context(), "b", false),
		)

		assert.NotEqual(t, leaseOfTheOnlyKeyUnder(t, "a"), leaseOfTheOnlyKeyUnder(t, "b"))
	})

	t.Run("ends when the ability is forgotten, releasing every lock and every leadership", func(t *testing.T) {
		theAbility := ability.UseETCD().WithTheEndpoints(server.Endpoint())
		alice := ability.TheAbilityOf(screenplay.ActorNamed("Alice").WhoCan(theAbility))
		require.NoError(t, alice.ToAcquireTheLock(t.Context(), "release", false))
		require.NoError(t, alice.ToCampaign(t.Context(), "backup-worker", "Alice"))

		require.NoError(t, theAbility.Forget())

		for _, name := range []string{"release", "backup-worker"} {
			response, err := client.Get(t.Context(), name+"/", clientv3.WithPrefix())
			require.NoError(t, err)
			assert.Empty(t, response.Kvs, "the keys of %q are gone with the session", name)
		}
	})

	t.Run("is forgotten with the locks and the elections it held", func(t *testing.T) {
		theAbility := ability.UseETCD().WithTheEndpoints(server.Endpoint())
		alice := ability.TheAbilityOf(screenplay.ActorNamed("Alice").WhoCan(theAbility))
		require.NoError(t, alice.ToAcquireTheLock(t.Context(), "forgotten", false))
		require.NoError(t, alice.ToCampaign(t.Context(), "forgotten-worker", "Alice"))

		require.NoError(t, theAbility.Forget())

		require.ErrorIs(t, alice.ToReleaseTheLock(t.Context(), "forgotten"), etcd.ErrLockNotHeld)
		require.ErrorIs(t, alice.ToResign(t.Context(), "forgotten-worker"), etcd.ErrNotCampaigning)
	})

	t.Run("is never opened by an ability that only uses keys", func(t *testing.T) {
		theAbility := ability.UseETCD().WithTheEndpoints(server.Endpoint())
		alice := ability.TheAbilityOf(screenplay.ActorNamed("Alice").WhoCan(theAbility))
		before, err := client.Leases(t.Context())
		require.NoError(t, err)

		require.NoError(t, alice.ToPut(t.Context(), "app/status", "ready", 0))
		require.NoError(t, theAbility.Forget())

		after, err := client.Leases(t.Context())
		require.NoError(t, err)
		assert.Len(t, after.Leases, len(before.Leases))
	})

	t.Run("fails to open when no endpoint was provided", func(t *testing.T) {
		alice := ability.TheAbilityOf(screenplay.ActorNamed("Alice").WhoCan(ability.UseETCD()))

		require.ErrorIs(t, alice.ToAcquireTheLock(t.Context(), "deployment", false), etcd.ErrNoEndpointProvided)
		require.ErrorIs(t, alice.ToCampaign(t.Context(), "primary-worker", "Alice"), etcd.ErrNoEndpointProvided)
		_, err := alice.ToRetrieveTheLeader(t.Context(), "primary-worker")
		require.ErrorIs(t, err, etcd.ErrNoEndpointProvided)
	})

	t.Run("fails to open when the actor does not have the ability", func(t *testing.T) {
		noAbility := ability.TheAbilityOf(screenplay.ActorNamed("Bob"))

		require.ErrorIs(
			t,
			noAbility.ToAcquireTheLock(t.Context(), "deployment", false),
			screenplay.ErrActorMissingAbility,
		)
		require.ErrorIs(
			t,
			noAbility.ToCampaign(t.Context(), "primary-worker", "Bob"),
			screenplay.ErrActorMissingAbility,
		)
		_, err := noAbility.ToRetrieveTheLeader(t.Context(), "primary-worker")
		require.ErrorIs(t, err, screenplay.ErrActorMissingAbility)
	})
}
