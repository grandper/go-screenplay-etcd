package events_test

import (
	"testing"
	"time"

	"github.com/grandper/go-screenplay/v2/action/see"
	"github.com/grandper/go-screenplay/v2/action/wait"
	"github.com/grandper/go-screenplay/v2/question/last"
	"github.com/grandper/go-screenplay/v2/resolution/length"
	"github.com/grandper/go-screenplay/v2/screenplay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grandper/go-screenplay-etcd/ability"
	"github.com/grandper/go-screenplay-etcd/action/put"
	"github.com/grandper/go-screenplay-etcd/action/remove"
	"github.com/grandper/go-screenplay-etcd/action/watch"
	"github.com/grandper/go-screenplay-etcd/fixture"
	"github.com/grandper/go-screenplay-etcd/question"
	"github.com/grandper/go-screenplay-etcd/question/events"
	"github.com/grandper/go-screenplay-etcd/resolution/been"
)

func TestOnTheKey(t *testing.T) {
	server := fixture.RunETCDServer(t)
	production := screenplay.NewProduction(screenplay.WithTimeout(15 * time.Second))
	alice := production.SetTheStage(screenplay.CastOfStandardActors()).
		ActorNamed("Alice").
		WhoCan(ability.UseETCD().WithTheEndpoints(server.Endpoint()))
	t.Cleanup(func() { require.NoError(t, alice.Exit()) })

	t.Run("answers with the events recorded on the key, in order", func(t *testing.T) {
		require.NoError(t, alice.AttemptsTo(
			watch.TheKey("app/config/database"),
			put.TheValue("postgresql://localhost:5432").ForTheKey("app/config/database"),
			remove.TheKey("app/config/database"),
			wait.UntilThe(events.OnTheKey("app/config/database")).
				HappensToHave(length.Exactly[[]*ability.Event](2)),
		))

		theEvents, err := alice.AsksFor(events.OnTheKey("app/config/database"))

		require.NoError(t, err)
		require.Len(t, theEvents, 2)
		assert.True(t, theEvents[0].IsPut())
		assert.True(t, theEvents[1].IsDelete())
	})

	t.Run("can be asserted on with the been resolutions", func(t *testing.T) {
		require.NoError(t, alice.Should(
			see.The(last.Of(events.OnTheKey("app/config/database"))).ShouldHave(been.Deleted()),
			see.The(last.Of(events.OnTheKey("app/config/database"))).ShouldNotHave(been.Put()),
		))
	})

	t.Run("answers with no event when nothing happened to the key yet", func(t *testing.T) {
		require.NoError(t, alice.AttemptsTo(watch.TheKey("app/config/cache")))

		theEvents, err := alice.AsksFor(events.OnTheKey("app/config/cache"))

		require.NoError(t, err)
		assert.NotNil(t, theEvents)
		assert.Empty(t, theEvents)
	})

	t.Run("fails when the key is not watched", func(t *testing.T) {
		theEvents, err := alice.AsksFor(events.OnTheKey("app/status"))

		require.ErrorIs(t, err, ability.ErrNotWatching)
		require.EqualError(t, err, `the key "app/status" is not watched`)
		assert.Nil(t, theEvents)
	})

	t.Run("fails when the key is not provided", func(t *testing.T) {
		_, err := alice.AsksFor(events.OnTheKey(""))

		require.ErrorIs(t, err, question.ErrKeyNotProvided)
	})

	t.Run("fails when the actor cannot use etcd", func(t *testing.T) {
		_, err := screenplay.ActorNamed("Bob").AsksFor(events.OnTheKey("app/config/database"))

		require.ErrorIs(t, err, screenplay.ErrActorMissingAbility)
	})

	t.Run("implements the stringer interface", func(t *testing.T) {
		assert.Equal(
			t,
			`events on the key "app/config/database"`,
			events.OnTheKey("app/config/database").String(),
		)
	})

	t.Run("implements the question interface", func(t *testing.T) {
		assert.Implements(
			t,
			(*screenplay.Question[[]*ability.Event])(nil),
			events.OnTheKey("app/config/database"),
		)
	})
}
