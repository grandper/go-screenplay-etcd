package question_test

import (
	"testing"

	"github.com/grandper/go-screenplay/v2/action/see"
	"github.com/grandper/go-screenplay/v2/resolution/be"
	"github.com/grandper/go-screenplay/v2/screenplay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grandper/go-screenplay-etcd/ability"
	"github.com/grandper/go-screenplay-etcd/action/put"
	"github.com/grandper/go-screenplay-etcd/fixture"
	"github.com/grandper/go-screenplay-etcd/question"
)

func TestNumberOfKeysStartingWith(t *testing.T) {
	server := fixture.RunETCDServer(t)
	alice := screenplay.ActorNamed("Alice").WhoCan(ability.UseETCD().WithTheEndpoints(server.Endpoint()))
	require.NoError(t, alice.AttemptsTo(
		put.TheValue("postgresql://localhost:5432").ForTheKey("app/config/database"),
		put.TheValue("redis://localhost:6379").ForTheKey("app/config/cache"),
		put.TheValue("ready").ForTheKey("app/status"),
	))

	t.Run("answers with the number of keys starting with the prefix", func(t *testing.T) {
		number, err := alice.AsksFor(question.NumberOfKeysStartingWith("app/config/"))

		require.NoError(t, err)
		assert.Equal(t, 2, number)
	})

	t.Run("answers with zero when no key starts with the prefix", func(t *testing.T) {
		number, err := alice.AsksFor(question.NumberOfKeysStartingWith("nothing/"))

		require.NoError(t, err)
		assert.Zero(t, number)
	})

	t.Run("can be asserted on", func(t *testing.T) {
		require.NoError(t, alice.Should(
			see.The(question.NumberOfKeysStartingWith("app/")).Should(be.EqualTo(3)),
			see.The(question.NumberOfKeysStartingWith("app/config/")).Should(be.GreaterThan(0)),
		))
	})

	t.Run("fails when the prefix is not provided", func(t *testing.T) {
		number, err := alice.AsksFor(question.NumberOfKeysStartingWith(""))

		require.ErrorIs(t, err, question.ErrPrefixNotProvided)
		assert.Zero(t, number)
	})

	t.Run("fails when the actor cannot use etcd", func(t *testing.T) {
		_, err := screenplay.ActorNamed("Bob").AsksFor(question.NumberOfKeysStartingWith("app/config/"))

		require.ErrorIs(t, err, screenplay.ErrActorMissingAbility)
	})

	t.Run("implements the stringer interface", func(t *testing.T) {
		assert.Equal(t,
			`number of keys starting with "app/config/"`,
			question.NumberOfKeysStartingWith("app/config/").String(),
		)
	})

	t.Run("implements the question interface", func(t *testing.T) {
		assert.Implements(t, (*screenplay.Question[int])(nil), question.NumberOfKeysStartingWith("app/config/"))
	})
}
