package ttl_test

import (
	"testing"
	"time"

	"github.com/grandper/go-screenplay/v2/action/see"
	"github.com/grandper/go-screenplay/v2/resolution/be"
	"github.com/grandper/go-screenplay/v2/screenplay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grandper/go-screenplay-etcd/ability/use"
	"github.com/grandper/go-screenplay-etcd/action/put"
	"github.com/grandper/go-screenplay-etcd/etcd"
	"github.com/grandper/go-screenplay-etcd/internal/fixture"
	"github.com/grandper/go-screenplay-etcd/question/ttl"
)

func TestOfTheKey(t *testing.T) {
	server := fixture.RunETCDServer(t)
	alice := screenplay.ActorNamed("Alice").WhoCan(use.ETCD().WithTheEndpoints(server.Endpoint()))

	t.Run("answers with the time the key has left to live", func(t *testing.T) {
		require.NoError(
			t,
			alice.AttemptsTo(put.TheValue("held").ForTheKey("locks/batch-job").ExpiringAfter(30).Seconds()),
		)

		timeToLive, err := alice.AsksFor(ttl.OfTheKey("locks/batch-job"))

		require.NoError(t, err)
		assert.Greater(t, timeToLive, 20*time.Second)
		assert.LessOrEqual(t, timeToLive, 30*time.Second)
		require.NoError(t, alice.Should(
			see.The(ttl.OfTheKey("locks/batch-job")).Should(be.AtMost(30*time.Second)),
		))
	})

	t.Run("answers with zero for a key that never expires", func(t *testing.T) {
		require.NoError(t, alice.AttemptsTo(put.TheValue("ready").ForTheKey("app/status")))

		timeToLive, err := alice.AsksFor(ttl.OfTheKey("app/status"))

		require.NoError(t, err)
		assert.Zero(t, timeToLive)
	})

	t.Run("fails when the key does not exist", func(t *testing.T) {
		_, err := alice.AsksFor(ttl.OfTheKey("does/not/exist"))

		require.ErrorIs(t, err, etcd.ErrKeyNotFound)
	})

	t.Run("fails when the key is not provided", func(t *testing.T) {
		_, err := alice.AsksFor(ttl.OfTheKey(""))

		require.ErrorIs(t, err, etcd.ErrKeyNotProvided)
	})

	t.Run("fails when the actor cannot use etcd", func(t *testing.T) {
		_, err := screenplay.ActorNamed("Bob").AsksFor(ttl.OfTheKey("app/status"))

		require.ErrorIs(t, err, screenplay.ErrActorMissingAbility)
	})

	t.Run("implements the stringer interface", func(t *testing.T) {
		assert.Equal(
			t,
			`TTL of the key "locks/batch-job"`,
			ttl.OfTheKey("locks/batch-job").String(),
		)
	})

	t.Run("implements the question interface", func(t *testing.T) {
		assert.Implements(t, (*screenplay.Question[time.Duration])(nil), ttl.OfTheKey("locks/batch-job"))
	})
}
