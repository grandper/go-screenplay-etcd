package etcd_test

import (
	"testing"

	"github.com/grandper/go-screenplay/v2/screenplay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grandper/go-screenplay-etcd/ability/use"
	"github.com/grandper/go-screenplay-etcd/action/put"
	"github.com/grandper/go-screenplay-etcd/internal/fixture"
	"github.com/grandper/go-screenplay-etcd/question/key"
)

func TestKey(t *testing.T) {
	server := fixture.RunETCDServer(t)
	alice := screenplay.ActorNamed("Alice").WhoCan(use.ETCD().WithTheEndpoints(server.Endpoint()))
	t.Cleanup(func() { require.NoError(t, alice.Exit()) })

	t.Run("exposes the parts of a key as the actor saw it", func(t *testing.T) {
		require.NoError(t, alice.AttemptsTo(
			put.TheValue("1").ForTheKey("app/config/retries"),
			put.TheValue("3").ForTheKey("app/config/retries"),
		))

		theKey, err := alice.AsksFor(key.Named("app/config/retries"))

		require.NoError(t, err)
		assert.True(t, theKey.Exists())
		assert.Equal(t, "app/config/retries", theKey.Name())
		assert.Equal(t, "3", theKey.Value())
		assert.Equal(t, int64(2), theKey.Version(), "the key was put twice")
		assert.Positive(t, theKey.CreateRevision())
		assert.Greater(t, theKey.ModRevision(), theKey.CreateRevision(), "the key was modified after its creation")
		assert.Zero(t, theKey.Lease(), "the key never expires")
	})

	t.Run("exposes the lease of a key that expires", func(t *testing.T) {
		require.NoError(
			t,
			alice.AttemptsTo(put.TheValue("held").ForTheKey("locks/batch-job").ExpiringAfter(30).Seconds()),
		)

		theKey, err := alice.AsksFor(key.Named("locks/batch-job"))

		require.NoError(t, err)
		assert.NotZero(t, theKey.Lease())
	})

	t.Run("exposes nothing but the name of a key that does not exist", func(t *testing.T) {
		theKey, err := alice.AsksFor(key.Named("does/not/exist"))

		require.NoError(t, err)
		assert.False(t, theKey.Exists())
		assert.Equal(t, "does/not/exist", theKey.Name())
		assert.Empty(t, theKey.Value())
		assert.Zero(t, theKey.Version())
		assert.Zero(t, theKey.CreateRevision())
		assert.Zero(t, theKey.ModRevision())
		assert.Zero(t, theKey.Lease())
	})
}
