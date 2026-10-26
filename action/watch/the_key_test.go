package watch_test

import (
	"testing"
	"time"

	"github.com/grandper/go-screenplay-etcd/question/events"

	"github.com/grandper/go-screenplay/v2/action/wait"
	"github.com/grandper/go-screenplay/v2/resolution/length"
	"github.com/grandper/go-screenplay/v2/screenplay"
	screenplayfixture "github.com/grandper/go-screenplay/v2/screenplay/fixture"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grandper/go-screenplay-etcd/action/put"
	"github.com/grandper/go-screenplay-etcd/action/remove"
	"github.com/grandper/go-screenplay-etcd/action/watch"
	"github.com/grandper/go-screenplay-etcd/etcd"
	"github.com/grandper/go-screenplay-etcd/fixture"
	etcdfixture "github.com/grandper/go-screenplay-etcd/internal/fixture"
	"github.com/grandper/go-screenplay-etcd/use"
)

func TestTheKey(t *testing.T) {
	server := fixture.RunETCDServer(t)
	production := screenplay.NewProduction(screenplay.WithTimeout(15 * time.Second))
	stage := production.SetTheStage(screenplay.CastOfStandardActors())
	bob := screenplay.ActorNamed("Bob").WhoCan(use.ETCD().WithTheEndpoints(server.Endpoint()))

	t.Run("watches a key", func(t *testing.T) {
		t.Run("and records the changes another actor makes to it", func(t *testing.T) {
			alice := stage.ActorNamed("Alice").WhoCan(use.ETCD().WithTheEndpoints(server.Endpoint()))
			t.Cleanup(func() { require.NoError(t, alice.Exit()) })

			require.NoError(t, alice.AttemptsTo(watch.TheKey("app/config/database")))
			require.NoError(t, bob.AttemptsTo(
				put.TheValue("postgresql://localhost:5432").ForTheKey("app/config/database"),
				put.TheValue("postgresql://db.example.com:5432").ForTheKey("app/config/database"),
				remove.TheKey("app/config/database"),
			))

			require.NoError(t, alice.AttemptsTo(
				wait.UntilThe(events.OnTheKey("app/config/database")).
					HappensToHave(length.Exactly[[]*etcd.Event](3)),
			))
			theEvents, err := alice.AsksFor(events.OnTheKey("app/config/database"))
			require.NoError(t, err)
			require.Len(t, theEvents, 3)

			assert.True(t, theEvents[0].IsPut())
			assert.Equal(t, "app/config/database", theEvents[0].Key())
			assert.Equal(t, "postgresql://localhost:5432", theEvents[0].Value())
			assert.Empty(t, theEvents[0].PreviousValue())
			assert.Equal(t, int64(1), theEvents[0].Version())

			assert.True(t, theEvents[1].IsPut())
			assert.Equal(t, "postgresql://db.example.com:5432", theEvents[1].Value())
			assert.Equal(t, "postgresql://localhost:5432", theEvents[1].PreviousValue())
			assert.Equal(t, int64(2), theEvents[1].Version())
			assert.Greater(t, theEvents[1].Revision(), theEvents[0].Revision())

			assert.True(t, theEvents[2].IsDelete())
			assert.False(t, theEvents[2].IsPut())
			assert.Empty(t, theEvents[2].Value())
			assert.Equal(t, "postgresql://db.example.com:5432", theEvents[2].PreviousValue())
			assert.False(t, theEvents[2].CausedByLeaseExpiry())
		})

		t.Run("ignoring the changes made to other keys", func(t *testing.T) {
			alice := stage.ActorNamed("Alice").WhoCan(use.ETCD().WithTheEndpoints(server.Endpoint()))
			t.Cleanup(func() { require.NoError(t, alice.Exit()) })

			require.NoError(t, alice.AttemptsTo(watch.TheKey("app/config/cache")))
			require.NoError(t, bob.AttemptsTo(
				put.TheValue("ready").ForTheKey("app/status"),
				put.TheValue("redis://localhost:6379").ForTheKey("app/config/cache"),
			))

			require.NoError(t, alice.AttemptsTo(
				wait.UntilThe(events.OnTheKey("app/config/cache")).
					HappensToHave(length.Exactly[[]*etcd.Event](1)),
			))
			theEvents, err := alice.AsksFor(events.OnTheKey("app/config/cache"))
			require.NoError(t, err)
			assert.Equal(t, "app/config/cache", theEvents[0].Key())
		})

		t.Run("telling a removal caused by the expiry of a lease from a plain removal", func(t *testing.T) {
			alice := stage.ActorNamed("Alice").WhoCan(use.ETCD().WithTheEndpoints(server.Endpoint()))
			t.Cleanup(func() { require.NoError(t, alice.Exit()) })

			require.NoError(t, alice.AttemptsTo(watch.TheKey("locks/batch-job"), watch.TheKey("locks/nightly-job")))
			require.NoError(t, bob.AttemptsTo(
				put.TheValue("held").ForTheKey("locks/batch-job").ExpiringAfter(1).Second(),
				put.TheValue("held").ForTheKey("locks/nightly-job").ExpiringAfter(30).Seconds(),
				remove.TheKey("locks/nightly-job"),
			))

			require.NoError(t, alice.AttemptsTo(
				wait.UntilThe(events.OnTheKey("locks/batch-job")).
					HappensToHave(length.Exactly[[]*etcd.Event](2)),
				wait.UntilThe(events.OnTheKey("locks/nightly-job")).
					HappensToHave(length.Exactly[[]*etcd.Event](2)),
			))
			expired, err := alice.AsksFor(events.OnTheKey("locks/batch-job"))
			require.NoError(t, err)
			assert.True(t, expired[1].IsDelete())
			assert.True(t, expired[1].CausedByLeaseExpiry())
			removed, err := alice.AsksFor(events.OnTheKey("locks/nightly-job"))
			require.NoError(t, err)
			assert.True(t, removed[1].IsDelete())
			assert.False(t, removed[1].CausedByLeaseExpiry())
		})

		t.Run("fails when the key is already watched", func(t *testing.T) {
			alice := stage.ActorNamed("Alice").WhoCan(use.ETCD().WithTheEndpoints(server.Endpoint()))
			t.Cleanup(func() { require.NoError(t, alice.Exit()) })
			require.NoError(t, alice.AttemptsTo(watch.TheKey("app/config/database")))

			err := alice.AttemptsTo(watch.TheKey("app/config/database"))

			require.ErrorIs(t, err, etcd.ErrAlreadyWatching)
			require.EqualError(t, err, `the key "app/config/database" is already watched`)
		})

		t.Run("fails when the key is not provided", func(t *testing.T) {
			require.ErrorIs(t, bob.AttemptsTo(watch.TheKey("")), etcd.ErrKeyNotProvided)
		})

		t.Run("fails when the actor cannot use etcd", func(t *testing.T) {
			err := screenplay.ActorNamed("Carol").AttemptsTo(watch.TheKey("app/config/database"))

			require.ErrorIs(t, err, screenplay.ErrActorMissingAbility)
		})
	})

	t.Run("stops watching when the actor exits", func(t *testing.T) {
		alice := stage.ActorNamed("Alice").WhoCan(use.ETCD().WithTheEndpoints(server.Endpoint()))
		require.NoError(t, alice.AttemptsTo(watch.TheKey("app/config/database")))

		require.NoError(t, alice.Exit())

		_, err := alice.AsksFor(events.OnTheKey("app/config/database"))
		require.ErrorIs(t, err, screenplay.ErrActorMissingAbility)
	})

	t.Run("is narrated", func(t *testing.T) {
		recorder := screenplayfixture.NewRecorder()
		adam := etcdfixture.NarratedActor("Adam", recorder, server.Endpoint())
		t.Cleanup(func() { require.NoError(t, adam.Exit()) })

		require.NoError(t, adam.AttemptsTo(watch.TheKey("app/config/database")))

		assert.Equal(t, `Adam attempts to watch the key "app/config/database"`, recorder.Messages()[0])
	})

	t.Run("implements the stringer interface", func(t *testing.T) {
		assert.Equal(t, `watch the key "app/config/database"`, watch.TheKey("app/config/database").String())
	})
}
