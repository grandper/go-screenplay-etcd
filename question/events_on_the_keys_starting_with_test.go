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
	"github.com/grandper/go-screenplay-etcd/action/watch"
	"github.com/grandper/go-screenplay-etcd/fixture"
	"github.com/grandper/go-screenplay-etcd/question"
	"github.com/grandper/go-screenplay-etcd/resolution/been"
)

func TestEventsOnTheKeysStartingWith(t *testing.T) {
	server := fixture.RunETCDServer(t)
	production := screenplay.NewProduction(screenplay.WithTimeout(15 * time.Second))
	alice := production.SetTheStage(screenplay.CastOfStandardActors()).
		ActorNamed("Alice").
		WhoCan(ability.UseETCD().WithTheEndpoints(server.Endpoint()))
	t.Cleanup(func() { require.NoError(t, alice.Exit()) })

	t.Run("answers with the events recorded on the keys starting with the prefix, in order", func(t *testing.T) {
		require.NoError(t, alice.AttemptsTo(
			watch.TheKeysStartingWith("app/config/"),
			put.TheValue("redis://localhost:6379").ForTheKey("app/config/cache"),
			put.TheValue("postgresql://localhost:5432").ForTheKey("app/config/database"),
			wait.UntilThe(question.EventsOnTheKeysStartingWith("app/config/")).
				HappensToHave(length.Exactly[[]*ability.Event](2)),
		))

		events, err := alice.AsksFor(question.EventsOnTheKeysStartingWith("app/config/"))

		require.NoError(t, err)
		require.Len(t, events, 2)
		assert.Equal(t, "app/config/cache", events[0].Key())
		assert.Equal(t, "app/config/database", events[1].Key())
		require.NoError(t, alice.Should(
			see.The(last.Of(question.EventsOnTheKeysStartingWith("app/config/"))).ShouldHave(been.Put()),
		))
	})

	t.Run("fails when the prefix is not watched", func(t *testing.T) {
		events, err := alice.AsksFor(question.EventsOnTheKeysStartingWith("app/"))

		require.ErrorIs(t, err, ability.ErrNotWatching)
		require.EqualError(t, err, `the keys starting with "app/" are not watched`)
		assert.Nil(t, events)
	})

	t.Run("fails when the prefix is not provided", func(t *testing.T) {
		_, err := alice.AsksFor(question.EventsOnTheKeysStartingWith(""))

		require.ErrorIs(t, err, question.ErrPrefixNotProvided)
	})

	t.Run("fails when the actor cannot use etcd", func(t *testing.T) {
		_, err := screenplay.ActorNamed("Bob").AsksFor(question.EventsOnTheKeysStartingWith("app/config/"))

		require.ErrorIs(t, err, screenplay.ErrActorMissingAbility)
	})

	t.Run("implements the stringer interface", func(t *testing.T) {
		assert.Equal(t,
			`events on the keys starting with "app/config/"`,
			question.EventsOnTheKeysStartingWith("app/config/").String(),
		)
	})

	t.Run("implements the question interface", func(t *testing.T) {
		assert.Implements(t,
			(*screenplay.Question[[]*ability.Event])(nil),
			question.EventsOnTheKeysStartingWith("app/config/"),
		)
	})
}
