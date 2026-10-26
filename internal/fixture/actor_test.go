package fixture_test

import (
	"testing"

	"github.com/grandper/go-screenplay/v2/screenplay"
	screenplayfixture "github.com/grandper/go-screenplay/v2/screenplay/fixture"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grandper/go-screenplay-etcd/action/put"
	"github.com/grandper/go-screenplay-etcd/action/remove"
	"github.com/grandper/go-screenplay-etcd/fixture"
	etcdfixture "github.com/grandper/go-screenplay-etcd/internal/fixture"
	"github.com/grandper/go-screenplay-etcd/use"
)

func TestNarratedActor(t *testing.T) {
	server := fixture.RunETCDServer(t)
	client := server.NewClient(t)

	t.Run("creates an actor whose narration is recorded", func(t *testing.T) {
		recorder := screenplayfixture.NewRecorder()
		adam := etcdfixture.NarratedActor("Adam", recorder, server.Endpoint())
		t.Cleanup(func() { require.NoError(t, adam.Exit()) })

		require.NoError(t, adam.AttemptsTo(put.TheValue("ready").ForTheKey("app/status")))

		assert.Equal(t, "Adam", adam.Name())
		require.NotEmpty(t, recorder.Messages())
		assert.Equal(t, `Adam attempts to put the value "ready" for the key "app/status"`, recorder.Messages()[0])
	})

	t.Run("creates an actor who can use the etcd server at the endpoint", func(t *testing.T) {
		adam := etcdfixture.NarratedActor("Adam", screenplayfixture.NewRecorder(), server.Endpoint())
		t.Cleanup(func() { require.NoError(t, adam.Exit()) })

		require.NoError(t, adam.AttemptsTo(put.TheValue("redis://localhost:6379").ForTheKey("app/config/cache")))

		assert.Equal(t, "redis://localhost:6379", etcdfixture.ValueOf(t, client, "app/config/cache"))
	})
}

func TestValueOf(t *testing.T) {
	server := fixture.RunETCDServer(t)
	client := server.NewClient(t)
	alice := screenplay.ActorNamed("Alice").WhoCan(use.ETCD().WithTheEndpoints(server.Endpoint()))

	t.Run("returns the value the key holds", func(t *testing.T) {
		require.NoError(
			t,
			alice.AttemptsTo(put.TheValue("postgresql://localhost:5432").ForTheKey("app/config/database")),
		)

		assert.Equal(t, "postgresql://localhost:5432", etcdfixture.ValueOf(t, client, "app/config/database"))
	})

	t.Run("returns an empty value when that is what the key holds", func(t *testing.T) {
		require.NoError(t, alice.AttemptsTo(put.TheValue("").ForTheKey("app/config/empty")))

		assert.Empty(t, etcdfixture.ValueOf(t, client, "app/config/empty"))
	})
}

func TestAssertKeyDoesNotExist(t *testing.T) {
	server := fixture.RunETCDServer(t)
	client := server.NewClient(t)
	alice := screenplay.ActorNamed("Alice").WhoCan(use.ETCD().WithTheEndpoints(server.Endpoint()))

	t.Run("passes for a key that was never put", func(t *testing.T) {
		etcdfixture.AssertKeyDoesNotExist(t, client, "does/not/exist")
	})

	t.Run("passes for a key that was removed", func(t *testing.T) {
		require.NoError(t, alice.AttemptsTo(
			put.TheValue("ready").ForTheKey("app/status"),
			remove.TheKey("app/status"),
		))

		etcdfixture.AssertKeyDoesNotExist(t, client, "app/status")
	})
}

func TestRevisionOf(t *testing.T) {
	server := fixture.RunETCDServer(t)
	client := server.NewClient(t)
	alice := screenplay.ActorNamed("Alice").WhoCan(use.ETCD().WithTheEndpoints(server.Endpoint()))

	t.Run("returns the revision of the store, which every write increments", func(t *testing.T) {
		before := etcdfixture.RevisionOf(t, client)

		require.NoError(t, alice.AttemptsTo(put.TheValue("ready").ForTheKey("app/status")))

		assert.Equal(t, before+1, etcdfixture.RevisionOf(t, client))
	})

	t.Run("returns the same revision while nothing is written", func(t *testing.T) {
		before := etcdfixture.RevisionOf(t, client)

		assert.Equal(t, before, etcdfixture.RevisionOf(t, client))
	})
}
