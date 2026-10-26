package values_test

import (
	"testing"

	"github.com/grandper/go-screenplay/v2/action/see"
	"github.com/grandper/go-screenplay/v2/resolution/be"
	"github.com/grandper/go-screenplay/v2/screenplay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grandper/go-screenplay-etcd/action/put"
	"github.com/grandper/go-screenplay-etcd/etcd"
	"github.com/grandper/go-screenplay-etcd/fixture"
	"github.com/grandper/go-screenplay-etcd/question/values"
	"github.com/grandper/go-screenplay-etcd/use"
)

func TestOfTheKeysStartingWith(t *testing.T) {
	server := fixture.RunETCDServer(t)
	alice := screenplay.ActorNamed("Alice").WhoCan(use.ETCD().WithTheEndpoints(server.Endpoint()))
	require.NoError(t, alice.AttemptsTo(
		put.TheValue("redis://localhost:6379").ForTheKey("app/config/cache"),
		put.TheValue("postgresql://localhost:5432").ForTheKey("app/config/database"),
		put.TheValue("ready").ForTheKey("app/status"),
	))

	t.Run("answers with the values of the keys starting with the prefix", func(t *testing.T) {
		theValues, err := alice.AsksFor(values.OfTheKeysStartingWith("app/config/"))

		require.NoError(t, err)
		assert.Equal(t, map[string]string{
			"app/config/cache":    "redis://localhost:6379",
			"app/config/database": "postgresql://localhost:5432",
		}, theValues)
	})

	t.Run("answers with an empty map when no key starts with the prefix", func(t *testing.T) {
		theValues, err := alice.AsksFor(values.OfTheKeysStartingWith("nothing/"))

		require.NoError(t, err)
		assert.NotNil(t, theValues)
		assert.Empty(t, theValues)
	})

	t.Run("can be asserted on", func(t *testing.T) {
		require.NoError(t, alice.Should(
			see.The(values.OfTheKeysStartingWith("app/config/")).Should(be.DeepEqualTo(map[string]string{
				"app/config/cache":    "redis://localhost:6379",
				"app/config/database": "postgresql://localhost:5432",
			})),
		))
		require.NoError(t, alice.Should(
			see.The(values.OfTheKeysStartingWith("nothing/")).Should(be.Empty[map[string]string]()),
		))
	})

	t.Run("fails when the prefix is not provided", func(t *testing.T) {
		theValues, err := alice.AsksFor(values.OfTheKeysStartingWith(""))

		require.ErrorIs(t, err, etcd.ErrPrefixNotProvided)
		assert.Nil(t, theValues)
	})

	t.Run("fails when the actor cannot use etcd", func(t *testing.T) {
		_, err := screenplay.ActorNamed("Bob").AsksFor(values.OfTheKeysStartingWith("app/config/"))

		require.ErrorIs(t, err, screenplay.ErrActorMissingAbility)
	})

	t.Run("implements the stringer interface", func(t *testing.T) {
		assert.Equal(t,
			`values of the keys starting with "app/config/"`,
			values.OfTheKeysStartingWith("app/config/").String(),
		)
	})

	t.Run("implements the question interface", func(t *testing.T) {
		assert.Implements(t,
			(*screenplay.Question[map[string]string])(nil),
			values.OfTheKeysStartingWith("app/config/"),
		)
	})
}
