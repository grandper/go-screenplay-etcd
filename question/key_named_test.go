package question_test

import (
	"testing"

	"github.com/grandper/go-screenplay/v2/action/see"
	"github.com/grandper/go-screenplay/v2/screenplay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grandper/go-screenplay-etcd/ability"
	"github.com/grandper/go-screenplay-etcd/action/put"
	"github.com/grandper/go-screenplay-etcd/fixture"
	"github.com/grandper/go-screenplay-etcd/question"
	"github.com/grandper/go-screenplay-etcd/resolution"
)

func TestKeyNamed(t *testing.T) {
	server := fixture.RunETCDServer(t)
	alice := screenplay.ActorNamed("Alice").WhoCan(ability.UseETCD().WithTheEndpoints(server.Endpoint()))

	t.Run("answers with an existing key and its metadata", func(t *testing.T) {
		require.NoError(t, alice.AttemptsTo(
			put.TheValue("1").ForTheKey("app/config/retries"),
			put.TheValue("3").ForTheKey("app/config/retries"),
		))

		key, err := alice.AsksFor(question.KeyNamed("app/config/retries"))

		require.NoError(t, err)
		assert.True(t, key.Exists())
		assert.Equal(t, "app/config/retries", key.Name())
		assert.Equal(t, "3", key.Value())
		assert.Equal(t, int64(2), key.Version())
		assert.Positive(t, key.CreateRevision())
		assert.Greater(t, key.ModRevision(), key.CreateRevision())
		assert.Zero(t, key.Lease())
	})

	t.Run("answers with a key that does not exist", func(t *testing.T) {
		key, err := alice.AsksFor(question.KeyNamed("does/not/exist"))

		require.NoError(t, err)
		assert.False(t, key.Exists())
		assert.Equal(t, "does/not/exist", key.Name())
		assert.Empty(t, key.Value())
	})

	t.Run("can be asserted on with the exist resolution", func(t *testing.T) {
		require.NoError(t, alice.AttemptsTo(put.TheValue("ready").ForTheKey("app/status")))

		require.NoError(t, alice.Should(see.The(question.KeyNamed("app/status")).Should(resolution.Exist())))
		require.NoError(t, alice.Should(see.The(question.KeyNamed("does/not/exist")).ShouldNot(resolution.Exist())))

		err := alice.Should(see.The(question.KeyNamed("does/not/exist")).Should(resolution.Exist()))
		require.EqualError(
			t,
			err,
			`Alice failed to see that the key named "does/not/exist" should exist: it does not exist`,
		)
	})

	t.Run("fails when the key is not provided", func(t *testing.T) {
		key, err := alice.AsksFor(question.KeyNamed(""))

		require.ErrorIs(t, err, question.ErrKeyNotProvided)
		assert.Nil(t, key)
	})

	t.Run("fails when the actor cannot use etcd", func(t *testing.T) {
		key, err := screenplay.ActorNamed("Bob").AsksFor(question.KeyNamed("app/status"))

		require.ErrorIs(t, err, screenplay.ErrActorMissingAbility)
		assert.Nil(t, key)
	})

	t.Run("implements the stringer interface", func(t *testing.T) {
		assert.Equal(t, `key named "app/status"`, question.KeyNamed("app/status").String())
	})

	t.Run("implements the question interface", func(t *testing.T) {
		assert.Implements(t, (*screenplay.Question[*ability.Key])(nil), question.KeyNamed("app/status"))
	})
}
