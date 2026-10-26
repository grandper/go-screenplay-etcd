package key_test

import (
	"testing"

	"github.com/grandper/go-screenplay-etcd/question/key"

	"github.com/grandper/go-screenplay-etcd/resolution/exist"

	"github.com/grandper/go-screenplay/v2/action/see"
	"github.com/grandper/go-screenplay/v2/screenplay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grandper/go-screenplay-etcd/action/put"
	"github.com/grandper/go-screenplay-etcd/etcd"
	"github.com/grandper/go-screenplay-etcd/fixture"
	"github.com/grandper/go-screenplay-etcd/use"
)

func TestNamed(t *testing.T) {
	server := fixture.RunETCDServer(t)
	alice := screenplay.ActorNamed("Alice").WhoCan(use.ETCD().WithTheEndpoints(server.Endpoint()))

	t.Run("answers with an existing key and its metadata", func(t *testing.T) {
		require.NoError(t, alice.AttemptsTo(
			put.TheValue("1").ForTheKey("app/config/retries"),
			put.TheValue("3").ForTheKey("app/config/retries"),
		))

		theKey, err := alice.AsksFor(key.Named("app/config/retries"))

		require.NoError(t, err)
		assert.True(t, theKey.Exists())
		assert.Equal(t, "app/config/retries", theKey.Name())
		assert.Equal(t, "3", theKey.Value())
		assert.Equal(t, int64(2), theKey.Version())
		assert.Positive(t, theKey.CreateRevision())
		assert.Greater(t, theKey.ModRevision(), theKey.CreateRevision())
		assert.Zero(t, theKey.Lease())
	})

	t.Run("answers with a key that does not exist", func(t *testing.T) {
		theKey, err := alice.AsksFor(key.Named("does/not/exist"))

		require.NoError(t, err)
		assert.False(t, theKey.Exists())
		assert.Equal(t, "does/not/exist", theKey.Name())
		assert.Empty(t, theKey.Value())
	})

	t.Run("can be asserted on with the exist resolution", func(t *testing.T) {
		require.NoError(t, alice.AttemptsTo(put.TheValue("ready").ForTheKey("app/status")))

		require.NoError(t, alice.Should(see.The(key.Named("app/status")).Should(exist.InETCD())))
		require.NoError(t, alice.Should(see.The(key.Named("does/not/exist")).ShouldNot(exist.InETCD())))

		err := alice.Should(see.The(key.Named("does/not/exist")).Should(exist.InETCD()))
		require.EqualError(
			t,
			err,
			`Alice failed to see that the key named "does/not/exist" should exist in etcd: it does not exist`,
		)
	})

	t.Run("fails when the key is not provided", func(t *testing.T) {
		theKey, err := alice.AsksFor(key.Named(""))

		require.ErrorIs(t, err, etcd.ErrKeyNotProvided)
		assert.Nil(t, theKey)
	})

	t.Run("fails when the actor cannot use etcd", func(t *testing.T) {
		theKey, err := screenplay.ActorNamed("Bob").AsksFor(key.Named("app/status"))

		require.ErrorIs(t, err, screenplay.ErrActorMissingAbility)
		assert.Nil(t, theKey)
	})

	t.Run("implements the stringer interface", func(t *testing.T) {
		assert.Equal(t, `key named "app/status"`, key.Named("app/status").String())
	})

	t.Run("implements the question interface", func(t *testing.T) {
		assert.Implements(t, (*screenplay.Question[*etcd.Key])(nil), key.Named("app/status"))
	})
}
