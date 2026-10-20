package question_test

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
	"github.com/grandper/go-screenplay-etcd/resolution/been"
)

func TestEventsOnTheKey(t *testing.T) {
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
			wait.UntilThe(question.EventsOnTheKey("app/config/database")).
				HappensToHave(length.Exactly[[]*ability.Event](2)),
		))

		events, err := alice.AsksFor(question.EventsOnTheKey("app/config/database"))

		require.NoError(t, err)
		require.Len(t, events, 2)
		assert.True(t, events[0].IsPut())
		assert.True(t, events[1].IsDelete())
	})

	t.Run("can be asserted on with the been resolutions", func(t *testing.T) {
		require.NoError(t, alice.Should(
			see.The(last.Of(question.EventsOnTheKey("app/config/database"))).ShouldHave(been.Deleted()),
			see.The(last.Of(question.EventsOnTheKey("app/config/database"))).ShouldNotHave(been.Put()),
		))
	})

	t.Run("answers with no event when nothing happened to the key yet", func(t *testing.T) {
		require.NoError(t, alice.AttemptsTo(watch.TheKey("app/config/cache")))

		events, err := alice.AsksFor(question.EventsOnTheKey("app/config/cache"))

		require.NoError(t, err)
		assert.NotNil(t, events)
		assert.Empty(t, events)
	})

	t.Run("fails when the key is not watched", func(t *testing.T) {
		events, err := alice.AsksFor(question.EventsOnTheKey("app/status"))

		require.ErrorIs(t, err, ability.ErrNotWatching)
		require.EqualError(t, err, `the key "app/status" is not watched`)
		assert.Nil(t, events)
	})

	t.Run("fails when the key is not provided", func(t *testing.T) {
		_, err := alice.AsksFor(question.EventsOnTheKey(""))

		require.ErrorIs(t, err, question.ErrKeyNotProvided)
	})

	t.Run("fails when the actor cannot use etcd", func(t *testing.T) {
		_, err := screenplay.ActorNamed("Bob").AsksFor(question.EventsOnTheKey("app/config/database"))

		require.ErrorIs(t, err, screenplay.ErrActorMissingAbility)
	})

	t.Run("implements the stringer interface", func(t *testing.T) {
		assert.Equal(
			t,
			`events on the key "app/config/database"`,
			question.EventsOnTheKey("app/config/database").String(),
		)
	})

	t.Run("implements the question interface", func(t *testing.T) {
		assert.Implements(
			t,
			(*screenplay.Question[[]*ability.Event])(nil),
			question.EventsOnTheKey("app/config/database"),
		)
	})
}
