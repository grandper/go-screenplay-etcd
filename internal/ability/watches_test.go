package ability_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/grandper/go-screenplay/v2/screenplay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/grandper/go-screenplay-etcd/etcd"
	"github.com/grandper/go-screenplay-etcd/internal/ability"
	"github.com/grandper/go-screenplay-etcd/internal/fixture"
)

func TestWatches(t *testing.T) {
	server := fixture.RunETCDServer(t)

	t.Run("records the events on a watched key until the ability is forgotten", func(t *testing.T) {
		theAbility := ability.UseETCD().WithTheEndpoints(server.Endpoint())
		operations := ability.TheAbilityOf(screenplay.ActorNamed("Alice").WhoCan(theAbility))

		require.NoError(t, operations.ToWatch(t.Context(), "app/status", false))
		require.NoError(t, operations.ToPut(t.Context(), "app/status", "ready", 0))

		assert.Eventually(t, func() bool {
			events, err := operations.ToRetrieveTheEvents("app/status", false)

			return err == nil && len(events) == 1
		}, 15*time.Second, 100*time.Millisecond)

		require.NoError(t, theAbility.Forget())
		_, err := operations.ToRetrieveTheEvents("app/status", false)
		require.ErrorIs(t, err, etcd.ErrNotWatching)
	})

	t.Run("records the events on the keys starting with a watched prefix", func(t *testing.T) {
		theAbility := ability.UseETCD().WithTheEndpoints(server.Endpoint())
		operations := ability.TheAbilityOf(screenplay.ActorNamed("Alice").WhoCan(theAbility))
		t.Cleanup(func() { require.NoError(t, theAbility.Forget()) })

		require.NoError(t, operations.ToWatch(t.Context(), "app/config/", true))
		require.NoError(t, operations.ToPut(t.Context(), "app/config/cache", "redis://localhost:6379", 0))

		assert.Eventually(t, func() bool {
			events, err := operations.ToRetrieveTheEvents("app/config/", true)

			return err == nil && len(events) == 1
		}, 15*time.Second, 100*time.Millisecond)

		_, err := operations.ToRetrieveTheEvents("app/config/", false)
		require.ErrorIs(t, err, etcd.ErrNotWatching, "a watch of the key is not a watch of the prefix")
	})

	t.Run("drops a watch the store did not confirm in time", func(t *testing.T) {
		theAbility := ability.UseETCD().WithTheEndpoints(server.Endpoint())
		operations := ability.TheAbilityOf(screenplay.ActorNamed("Alice").WhoCan(theAbility))
		t.Cleanup(func() { require.NoError(t, theAbility.Forget()) })
		require.NoError(t, operations.ToPut(t.Context(), "app/status", "ready", 0), "connects beforehand")
		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		err := operations.ToWatch(ctx, "app/status", false)

		require.ErrorIs(t, err, context.Canceled)
		_, err = operations.ToRetrieveTheEvents("app/status", false)
		require.ErrorIs(t, err, etcd.ErrNotWatching, "the watch that was given up is not registered")
		require.NoError(t, operations.ToWatch(t.Context(), "app/status", false), "the watch can be attempted again")
	})

	t.Run("fails to answer the events once the store canceled the watch", func(t *testing.T) {
		client, err := clientv3.New(clientv3.Config{Endpoints: []string{server.Endpoint()}})
		require.NoError(t, err)
		theAbility := ability.UseETCD().WithTheClient(client)
		operations := ability.TheAbilityOf(screenplay.ActorNamed("Alice").WhoCan(theAbility))
		t.Cleanup(func() { require.NoError(t, theAbility.Forget()) })
		require.NoError(t, operations.ToWatch(t.Context(), "app/status", false))

		// Closing the watcher of the client ends every watch, as the store
		// does when it cancels one.
		require.NoError(t, client.Watcher.Close())

		assert.Eventually(t, func() bool {
			_, err = operations.ToRetrieveTheEvents("app/status", false)

			return errors.Is(err, etcd.ErrFailedToConnect)
		}, 15*time.Second, 100*time.Millisecond)
		require.ErrorContains(t, err, `the watch of the key "app/status" was canceled`)
	})

	t.Run("refuses to watch a key twice", func(t *testing.T) {
		theAbility := ability.UseETCD().WithTheEndpoints(server.Endpoint())
		operations := ability.TheAbilityOf(screenplay.ActorNamed("Alice").WhoCan(theAbility))
		t.Cleanup(func() { require.NoError(t, theAbility.Forget()) })
		require.NoError(t, operations.ToWatch(t.Context(), "app/status", false))

		require.ErrorIs(t, operations.ToWatch(t.Context(), "app/status", false), etcd.ErrAlreadyWatching)
		require.NoError(t, operations.ToWatch(t.Context(), "app/status", true), "a prefix is another watch")
	})

	t.Run("fails every operation when the actor does not have the ability", func(t *testing.T) {
		noAbility := ability.TheAbilityOf(screenplay.ActorNamed("Bob"))

		require.ErrorIs(t, noAbility.ToWatch(t.Context(), "k", false), screenplay.ErrActorMissingAbility)
		_, err := noAbility.ToRetrieveTheEvents("k", false)
		require.ErrorIs(t, err, screenplay.ErrActorMissingAbility)
	})
}
