package question_test

import (
	"testing"

	"github.com/grandper/go-screenplay/v2/action/see"
	"github.com/grandper/go-screenplay/v2/resolution/contain"
	"github.com/grandper/go-screenplay/v2/resolution/length"
	"github.com/grandper/go-screenplay/v2/screenplay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grandper/go-screenplay-etcd/ability"
	"github.com/grandper/go-screenplay-etcd/action/put"
	"github.com/grandper/go-screenplay-etcd/fixture"
	"github.com/grandper/go-screenplay-etcd/question"
)

func TestKeysStartingWith(t *testing.T) {
	server := fixture.RunETCDServer(t)
	alice := screenplay.ActorNamed("Alice").WhoCan(ability.UseETCD().WithTheEndpoints(server.Endpoint()))
	require.NoError(t, alice.AttemptsTo(
		put.TheValue("postgresql://localhost:5432").ForTheKey("app/config/database"),
		put.TheValue("redis://localhost:6379").ForTheKey("app/config/cache"),
		put.TheValue("ready").ForTheKey("app/status"),
	))

	t.Run("answers with the keys starting with the prefix, in order", func(t *testing.T) {
		keys, err := alice.AsksFor(question.KeysStartingWith("app/config/"))

		require.NoError(t, err)
		assert.Equal(t, []string{"app/config/cache", "app/config/database"}, keys)
	})

	t.Run("answers with an empty slice when no key starts with the prefix", func(t *testing.T) {
		keys, err := alice.AsksFor(question.KeysStartingWith("nothing/"))

		require.NoError(t, err)
		assert.NotNil(t, keys)
		assert.Empty(t, keys)
	})

	t.Run("can be asserted on", func(t *testing.T) {
		require.NoError(t, alice.Should(
			see.The(question.KeysStartingWith("app/")).ShouldHave(length.Exactly[[]string](3)),
			see.The(question.KeysStartingWith("app/")).Should(contain.Element("app/status")),
		))
	})

	t.Run("fails when the prefix is not provided", func(t *testing.T) {
		keys, err := alice.AsksFor(question.KeysStartingWith(""))

		require.ErrorIs(t, err, question.ErrPrefixNotProvided)
		assert.Nil(t, keys)
	})

	t.Run("fails when the actor cannot use etcd", func(t *testing.T) {
		_, err := screenplay.ActorNamed("Bob").AsksFor(question.KeysStartingWith("app/config/"))

		require.ErrorIs(t, err, screenplay.ErrActorMissingAbility)
	})

	t.Run("implements the stringer interface", func(t *testing.T) {
		assert.Equal(t, `keys starting with "app/config/"`, question.KeysStartingWith("app/config/").String())
	})

	t.Run("implements the question interface", func(t *testing.T) {
		assert.Implements(t, (*screenplay.Question[[]string])(nil), question.KeysStartingWith("app/config/"))
	})
}
