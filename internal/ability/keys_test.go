package ability_test

import (
	"testing"

	"github.com/grandper/go-screenplay/v2/screenplay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grandper/go-screenplay-etcd/internal/ability"
	"github.com/grandper/go-screenplay-etcd/internal/fixture"
)

func TestKeys(t *testing.T) {
	server := fixture.RunETCDServer(t)
	theAbility := ability.TheAbilityOf(
		screenplay.ActorNamed("Alice").WhoCan(ability.UseETCD().WithTheEndpoints(server.Endpoint())),
	)

	t.Run("puts and retrieves a key", func(t *testing.T) {
		require.NoError(t, theAbility.ToPut(t.Context(), "app/config/database", "postgresql://localhost:5432", 0))

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
		require.NoError(t, theAbility.ToPut(t.Context(), "app/config/cache", "redis://localhost:6379", 0))

		require.NoError(t, theAbility.ToRemove(t.Context(), "app/config/cache"))

		key, err := theAbility.ToRetrieveTheKey(t.Context(), "app/config/cache")
		require.NoError(t, err)
		assert.False(t, key.Exists())
	})

	t.Run("removes several keys in one transaction", func(t *testing.T) {
		require.NoError(t, theAbility.ToPut(t.Context(), "app/config/cache", "redis://localhost:6379", 0))
		require.NoError(t, theAbility.ToPut(t.Context(), "app/config/database", "postgresql://localhost:5432", 0))

		require.NoError(t, theAbility.ToRemoveTheKeys(t.Context(), "app/config/cache", "app/config/database"))

		for _, name := range []string{"app/config/cache", "app/config/database"} {
			key, err := theAbility.ToRetrieveTheKey(t.Context(), name)
			require.NoError(t, err)
			assert.False(t, key.Exists())
		}
	})

	t.Run("removes the keys starting with a prefix", func(t *testing.T) {
		require.NoError(t, theAbility.ToPut(t.Context(), "app/config/cache", "redis://localhost:6379", 0))
		require.NoError(t, theAbility.ToPut(t.Context(), "app/status", "ready", 0))

		require.NoError(t, theAbility.ToRemoveTheKeysStartingWith(t.Context(), "app/config/"))

		cache, err := theAbility.ToRetrieveTheKey(t.Context(), "app/config/cache")
		require.NoError(t, err)
		assert.False(t, cache.Exists())
		status, err := theAbility.ToRetrieveTheKey(t.Context(), "app/status")
		require.NoError(t, err)
		assert.True(t, status.Exists())
	})

	t.Run("retrieves the keys starting with a prefix, in order", func(t *testing.T) {
		require.NoError(t, theAbility.ToRemoveTheKeysStartingWith(t.Context(), "app/"))
		require.NoError(t, theAbility.ToPut(t.Context(), "app/config/database", "postgresql://localhost:5432", 0))
		require.NoError(t, theAbility.ToPut(t.Context(), "app/config/cache", "redis://localhost:6379", 0))
		require.NoError(t, theAbility.ToPut(t.Context(), "app/status", "ready", 0))

		keys, err := theAbility.ToRetrieveTheKeysStartingWith(t.Context(), "app/config/")

		require.NoError(t, err)
		require.Len(t, keys, 2)
		assert.Equal(t, "app/config/cache", keys[0].Name())
		assert.Equal(t, "redis://localhost:6379", keys[0].Value())
		assert.Equal(t, "app/config/database", keys[1].Name())
		assert.Equal(t, "postgresql://localhost:5432", keys[1].Value())

		none, err := theAbility.ToRetrieveTheKeysStartingWith(t.Context(), "nothing/")
		require.NoError(t, err)
		assert.Empty(t, none)
	})

	t.Run("puts a value only if the key is absent", func(t *testing.T) {
		require.NoError(t, theAbility.ToRemove(t.Context(), "app/config/flag"))

		succeeded, err := theAbility.ToPutIfAbsent(t.Context(), "app/config/flag", "initial", 0)
		require.NoError(t, err)
		assert.True(t, succeeded)

		succeeded, err = theAbility.ToPutIfAbsent(t.Context(), "app/config/flag", "other", 0)
		require.NoError(t, err)
		assert.False(t, succeeded)
		key, err := theAbility.ToRetrieveTheKey(t.Context(), "app/config/flag")
		require.NoError(t, err)
		assert.Equal(t, "initial", key.Value())
	})

	t.Run("puts a value only if the key is unchanged since a revision", func(t *testing.T) {
		require.NoError(t, theAbility.ToPut(t.Context(), "app/config/counter", "1", 0))
		key, err := theAbility.ToRetrieveTheKey(t.Context(), "app/config/counter")
		require.NoError(t, err)

		succeeded, err := theAbility.ToPutIfUnchanged(t.Context(), "app/config/counter", "2", key.ModRevision(), 0)
		require.NoError(t, err)
		assert.True(t, succeeded)

		succeeded, err = theAbility.ToPutIfUnchanged(t.Context(), "app/config/counter", "3", key.ModRevision(), 0)
		require.NoError(t, err)
		assert.False(t, succeeded)
	})

	t.Run("removes a key only if it is unchanged since a revision", func(t *testing.T) {
		require.NoError(t, theAbility.ToPut(t.Context(), "app/config/counter", "1", 0))
		key, err := theAbility.ToRetrieveTheKey(t.Context(), "app/config/counter")
		require.NoError(t, err)

		succeeded, err := theAbility.ToRemoveIfUnchanged(t.Context(), "app/config/counter", key.ModRevision()-1)
		require.NoError(t, err)
		assert.False(t, succeeded)

		succeeded, err = theAbility.ToRemoveIfUnchanged(t.Context(), "app/config/counter", key.ModRevision())
		require.NoError(t, err)
		assert.True(t, succeeded)
	})

	t.Run("puts several values in one transaction", func(t *testing.T) {
		require.NoError(t, theAbility.ToRemoveTheKeysStartingWith(t.Context(), "app/"))

		require.NoError(t, theAbility.ToPutTheValues(t.Context(), map[string]string{
			"app/config/database": "postgresql://localhost:5432",
			"app/config/cache":    "redis://localhost:6379",
		}))

		keys, err := theAbility.ToRetrieveTheKeysStartingWith(t.Context(), "app/config/")
		require.NoError(t, err)
		require.Len(t, keys, 2)
		assert.Equal(t, keys[0].ModRevision(), keys[1].ModRevision())
	})

	t.Run("fails every operation when the actor does not have the ability", func(t *testing.T) {
		noAbility := ability.TheAbilityOf(screenplay.ActorNamed("Bob"))

		require.ErrorIs(t, noAbility.ToPut(t.Context(), "k", "v", 0), screenplay.ErrActorMissingAbility)
		key, err := noAbility.ToRetrieveTheKey(t.Context(), "k")
		require.ErrorIs(t, err, screenplay.ErrActorMissingAbility)
		assert.Nil(t, key)
		require.ErrorIs(t, noAbility.ToRemove(t.Context(), "k"), screenplay.ErrActorMissingAbility)
		require.ErrorIs(t, noAbility.ToRemoveTheKeys(t.Context(), "k"), screenplay.ErrActorMissingAbility)
		require.ErrorIs(t, noAbility.ToRemoveTheKeysStartingWith(t.Context(), "k"), screenplay.ErrActorMissingAbility)
		keys, err := noAbility.ToRetrieveTheKeysStartingWith(t.Context(), "k")
		require.ErrorIs(t, err, screenplay.ErrActorMissingAbility)
		assert.Nil(t, keys)
		_, err = noAbility.ToPutIfAbsent(t.Context(), "k", "v", 0)
		require.ErrorIs(t, err, screenplay.ErrActorMissingAbility)
		_, err = noAbility.ToPutIfUnchanged(t.Context(), "k", "v", 1, 0)
		require.ErrorIs(t, err, screenplay.ErrActorMissingAbility)
		_, err = noAbility.ToRemoveIfUnchanged(t.Context(), "k", 1)
		require.ErrorIs(t, err, screenplay.ErrActorMissingAbility)
		require.ErrorIs(
			t,
			noAbility.ToPutTheValues(t.Context(), map[string]string{"k": "v"}),
			screenplay.ErrActorMissingAbility,
		)
	})
}
