package ability_test

import (
	"sync"
	"testing"

	"github.com/grandper/go-screenplay/v2/screenplay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grandper/go-screenplay-etcd/fixture"
	"github.com/grandper/go-screenplay-etcd/internal/ability"
)

func TestUseETCDAbility(t *testing.T) {
	server := fixture.RunETCDServer(t)

	t.Run("puts and retrieves a key", func(t *testing.T) {
		theAbility := ability.TheAbilityOf(
			screenplay.ActorNamed("Alice").WhoCan(ability.UseETCD().WithTheEndpoints(server.Endpoint())),
		)

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
		theAbility := ability.TheAbilityOf(
			screenplay.ActorNamed("Alice").WhoCan(ability.UseETCD().WithTheEndpoints(server.Endpoint())),
		)

		key, err := theAbility.ToRetrieveTheKey(t.Context(), "does/not/exist")
		require.NoError(t, err)
		assert.False(t, key.Exists())
		assert.Equal(t, "does/not/exist", key.Name())
		assert.Empty(t, key.Value())
		assert.Zero(t, key.Version())
	})

	t.Run("is safe for concurrent use", func(t *testing.T) {
		theAbility := ability.TheAbilityOf(
			screenplay.ActorNamed("Alice").WhoCan(ability.UseETCD().WithTheEndpoints(server.Endpoint())),
		)

		const branches = 10
		errs := make(chan error, 2*branches)
		var group sync.WaitGroup
		for i := range branches {
			group.Go(func() {
				errs <- theAbility.ToPut(t.Context(), "concurrent/key", string(rune('a'+i)))
				_, err := theAbility.ToRetrieveTheKey(t.Context(), "concurrent/key")
				errs <- err
			})
		}
		group.Wait()
		close(errs)

		for err := range errs {
			require.NoError(t, err)
		}
	})

	t.Run("fails every operation when the actor does not have the ability", func(t *testing.T) {
		theAbility := ability.TheAbilityOf(screenplay.ActorNamed("Bob"))

		require.ErrorIs(t, theAbility.ToPut(t.Context(), "k", "v"), screenplay.ErrActorMissingAbility)
		key, err := theAbility.ToRetrieveTheKey(t.Context(), "k")
		require.ErrorIs(t, err, screenplay.ErrActorMissingAbility)
		assert.Nil(t, key)
	})

	t.Run("fails every operation when the actor is nil", func(t *testing.T) {
		theAbility := ability.TheAbilityOf(nil)

		require.ErrorIs(t, theAbility.ToPut(t.Context(), "k", "v"), screenplay.ErrActorMissingAbility)
	})

	t.Run("reports the ability that was not given on a nil receiver", func(t *testing.T) {
		var theAbility *ability.UseETCDAbility

		require.ErrorIs(t, theAbility.Forget(), screenplay.ErrActorMissingAbility)
		assert.Nil(t, theAbility.WithTheEndpoints("localhost:2379"))
		assert.Nil(t, theAbility.WithTheNamespace("ns/"))
		assert.Nil(t, theAbility.WithTheClient(nil))
	})

	t.Run("can be forgotten more than once", func(t *testing.T) {
		theAbility := ability.UseETCD().WithTheEndpoints(server.Endpoint())
		require.NoError(
			t,
			ability.TheAbilityOf(screenplay.ActorNamed("Alice").WhoCan(theAbility)).ToPut(t.Context(), "k", "v"),
		)

		require.NoError(t, theAbility.Forget())
		require.NoError(t, theAbility.Forget())
	})

	t.Run("implements the stringer interface", func(t *testing.T) {
		assert.Equal(t, "use etcd", ability.UseETCD().String())
	})
}
