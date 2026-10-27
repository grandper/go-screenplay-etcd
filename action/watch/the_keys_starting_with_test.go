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

	"github.com/grandper/go-screenplay-etcd/ability/use"
	"github.com/grandper/go-screenplay-etcd/action/put"
	"github.com/grandper/go-screenplay-etcd/action/watch"
	"github.com/grandper/go-screenplay-etcd/etcd"
	"github.com/grandper/go-screenplay-etcd/internal/fixture"
)

func TestTheKeysStartingWith(t *testing.T) {
	server := fixture.RunETCDServer(t)
	production := screenplay.NewProduction(screenplay.WithTimeout(15 * time.Second))
	stage := production.SetTheStage(screenplay.CastOfStandardActors())
	bob := screenplay.ActorNamed("Bob").WhoCan(use.ETCD().WithTheEndpoints(server.Endpoint()))

	t.Run("watches the keys starting with a prefix", func(t *testing.T) {
		t.Run("and records the changes made to any of them", func(t *testing.T) {
			alice := stage.ActorNamed("Alice").WhoCan(use.ETCD().WithTheEndpoints(server.Endpoint()))
			t.Cleanup(func() { require.NoError(t, alice.Exit()) })

			require.NoError(t, alice.AttemptsTo(watch.TheKeysStartingWith("app/config/")))
			require.NoError(t, bob.AttemptsTo(
				put.TheValue("ready").ForTheKey("app/status"),
				put.TheValue("redis://localhost:6379").ForTheKey("app/config/cache"),
				put.TheValue("postgresql://localhost:5432").ForTheKey("app/config/database"),
			))

			require.NoError(t, alice.AttemptsTo(
				wait.UntilThe(events.OnTheKeysStartingWith("app/config/")).
					HappensToHave(length.Exactly[[]*etcd.Event](2)),
			))
			theEvents, err := alice.AsksFor(events.OnTheKeysStartingWith("app/config/"))
			require.NoError(t, err)
			assert.Equal(t, "app/config/cache", theEvents[0].Key())
			assert.Equal(t, "app/config/database", theEvents[1].Key())
		})

		t.Run("fails when the prefix is already watched", func(t *testing.T) {
			alice := stage.ActorNamed("Alice").WhoCan(use.ETCD().WithTheEndpoints(server.Endpoint()))
			t.Cleanup(func() { require.NoError(t, alice.Exit()) })
			require.NoError(t, alice.AttemptsTo(watch.TheKeysStartingWith("app/config/")))

			err := alice.AttemptsTo(watch.TheKeysStartingWith("app/config/"))

			require.ErrorIs(t, err, etcd.ErrAlreadyWatching)
			require.EqualError(t, err, `the keys starting with "app/config/" are already watched`)
		})

		t.Run("can be combined with a watch of one of the keys", func(t *testing.T) {
			alice := stage.ActorNamed("Alice").WhoCan(use.ETCD().WithTheEndpoints(server.Endpoint()))
			t.Cleanup(func() { require.NoError(t, alice.Exit()) })

			require.NoError(t, alice.AttemptsTo(
				watch.TheKeysStartingWith("app/config/"),
				watch.TheKey("app/config/"),
			))
		})

		t.Run("fails when the prefix is not provided", func(t *testing.T) {
			require.ErrorIs(t, bob.AttemptsTo(watch.TheKeysStartingWith("")), etcd.ErrPrefixNotProvided)
		})

		t.Run("fails when the actor cannot use etcd", func(t *testing.T) {
			err := screenplay.ActorNamed("Carol").AttemptsTo(watch.TheKeysStartingWith("app/config/"))

			require.ErrorIs(t, err, screenplay.ErrActorMissingAbility)
		})
	})

	t.Run("is narrated", func(t *testing.T) {
		recorder := screenplayfixture.NewRecorder()
		adam := fixture.NarratedActor("Adam", recorder, server.Endpoint())
		t.Cleanup(func() { require.NoError(t, adam.Exit()) })

		require.NoError(t, adam.AttemptsTo(watch.TheKeysStartingWith("app/config/")))

		assert.Equal(t, `Adam attempts to watch the keys starting with "app/config/"`, recorder.Messages()[0])
	})

	t.Run("implements the stringer interface", func(t *testing.T) {
		assert.Equal(t, `watch the keys starting with "app/config/"`, watch.TheKeysStartingWith("app/config/").String())
	})
}
