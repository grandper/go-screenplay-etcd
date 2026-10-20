package ability_test

import (
	"testing"

	"github.com/grandper/go-screenplay/v2/screenplay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grandper/go-screenplay-etcd/fixture"
	"github.com/grandper/go-screenplay-etcd/internal/ability"
)

func TestKeys(t *testing.T) {
	server := fixture.RunETCDServer(t)
	theAbility := ability.TheAbilityOf(
		screenplay.ActorNamed("Alice").WhoCan(ability.UseETCD().WithTheEndpoints(server.Endpoint())),
	)

	t.Run("puts and retrieves a key", func(t *testing.T) {
		require.NoError(t, theAbility.ToPut(t.Context(), "app/config/database", "postgresql://localhost:5432"))

		key, err := theAbility.ToRetrieveTheKey(t.Context(), "app/config/database")
		require.NoError(t, err)
		assert.True(t, key.Exists())
		assert.Equal(t, "app/config/database", key.Name())
		assert.Equal(t, "postgresql://localhost:5432", key.Value())
		assert.Equal(t, int64(1), key.Version())
		assert.Positive(t, key.CreateRevision())
		assert.Equal(t, key.CreateRevision(), key.ModRevision())
		assert.Zero(t, key.Lease())
	})

	t.Run("retrieves a key that does not exist", func(t *testing.T) {
		key, err := theAbility.ToRetrieveTheKey(t.Context(), "does/not/exist")

		require.NoError(t, err)
		assert.False(t, key.Exists())
		assert.Equal(t, "does/not/exist", key.Name())
		assert.Empty(t, key.Value())
		assert.Zero(t, key.Version())
	})

	t.Run("removes a key", func(t *testing.T) {
		require.NoError(t, theAbility.ToPut(t.Context(), "app/config/cache", "redis://localhost:6379"))

		require.NoError(t, theAbility.ToRemove(t.Context(), "app/config/cache"))

		key, err := theAbility.ToRetrieveTheKey(t.Context(), "app/config/cache")
		require.NoError(t, err)
		assert.False(t, key.Exists())
	})

	t.Run("removes several keys in one transaction", func(t *testing.T) {
		require.NoError(t, theAbility.ToPut(t.Context(), "app/config/cache", "redis://localhost:6379"))
		require.NoError(t, theAbility.ToPut(t.Context(), "app/config/database", "postgresql://localhost:5432"))

		require.NoError(t, theAbility.ToRemoveTheKeys(t.Context(), "app/config/cache", "app/config/database"))

		for _, name := range []string{"app/config/cache", "app/config/database"} {
			key, err := theAbility.ToRetrieveTheKey(t.Context(), name)
			require.NoError(t, err)
			assert.False(t, key.Exists())
		}
	})

	t.Run("removes the keys starting with a prefix", func(t *testing.T) {
		require.NoError(t, theAbility.ToPut(t.Context(), "app/config/cache", "redis://localhost:6379"))
		require.NoError(t, theAbility.ToPut(t.Context(), "app/status", "ready"))

		require.NoError(t, theAbility.ToRemoveTheKeysStartingWith(t.Context(), "app/config/"))

		cache, err := theAbility.ToRetrieveTheKey(t.Context(), "app/config/cache")
		require.NoError(t, err)
		assert.False(t, cache.Exists())
		status, err := theAbility.ToRetrieveTheKey(t.Context(), "app/status")
		require.NoError(t, err)
		assert.True(t, status.Exists())
	})

	t.Run("fails every operation when the actor does not have the ability", func(t *testing.T) {
		noAbility := ability.TheAbilityOf(screenplay.ActorNamed("Bob"))

		require.ErrorIs(t, noAbility.ToPut(t.Context(), "k", "v"), screenplay.ErrActorMissingAbility)
		key, err := noAbility.ToRetrieveTheKey(t.Context(), "k")
		require.ErrorIs(t, err, screenplay.ErrActorMissingAbility)
		assert.Nil(t, key)
		require.ErrorIs(t, noAbility.ToRemove(t.Context(), "k"), screenplay.ErrActorMissingAbility)
		require.ErrorIs(t, noAbility.ToRemoveTheKeys(t.Context(), "k"), screenplay.ErrActorMissingAbility)
		require.ErrorIs(t, noAbility.ToRemoveTheKeysStartingWith(t.Context(), "k"), screenplay.ErrActorMissingAbility)
	})
}
