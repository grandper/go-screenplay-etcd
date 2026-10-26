package question_test

import (
	"testing"

	"github.com/grandper/go-screenplay/v2/screenplay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grandper/go-screenplay-etcd/action/put"
	"github.com/grandper/go-screenplay-etcd/etcd"
	"github.com/grandper/go-screenplay-etcd/fixture"
	impl "github.com/grandper/go-screenplay-etcd/internal/question"
	"github.com/grandper/go-screenplay-etcd/use"
)

func TestRetrieveTheExistingKey(t *testing.T) {
	server := fixture.RunETCDServer(t)
	alice := screenplay.ActorNamed("Alice").WhoCan(use.ETCD().WithTheEndpoints(server.Endpoint()))

	t.Run("retrieves a key that exists", func(t *testing.T) {
		require.NoError(t, alice.AttemptsTo(put.TheValue("ready").ForTheKey("app/status")))

		key, err := impl.RetrieveTheExistingKey(alice, "app/status")

		require.NoError(t, err)
		assert.True(t, key.Exists())
		assert.Equal(t, "app/status", key.Name())
		assert.Equal(t, "ready", key.Value())
	})

	t.Run("fails when the key does not exist", func(t *testing.T) {
		key, err := impl.RetrieveTheExistingKey(alice, "does/not/exist")

		require.ErrorIs(t, err, etcd.ErrKeyNotFound)
		require.EqualError(t, err, `the key "does/not/exist" was not found`)
		assert.Nil(t, key)
	})

	t.Run("reports a missing key before reaching the store", func(t *testing.T) {
		key, err := impl.RetrieveTheExistingKey(screenplay.ActorNamed("Bob"), "")

		require.ErrorIs(t, err, etcd.ErrKeyNotProvided)
		require.NotErrorIs(t, err, screenplay.ErrActorMissingAbility)
		assert.Nil(t, key)
	})

	t.Run("fails when the actor cannot use etcd", func(t *testing.T) {
		key, err := impl.RetrieveTheExistingKey(screenplay.ActorNamed("Bob"), "app/status")

		require.ErrorIs(t, err, screenplay.ErrActorMissingAbility)
		assert.Nil(t, key)
	})
}
