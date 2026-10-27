package value_test

import (
	"testing"

	"github.com/grandper/go-screenplay/v2/action/see"
	"github.com/grandper/go-screenplay/v2/resolution/be"
	"github.com/grandper/go-screenplay/v2/screenplay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grandper/go-screenplay-etcd/ability/use"
	"github.com/grandper/go-screenplay-etcd/action/put"
	"github.com/grandper/go-screenplay-etcd/etcd"
	"github.com/grandper/go-screenplay-etcd/internal/fixture"
	"github.com/grandper/go-screenplay-etcd/question/value"
)

func TestOfTheKey(t *testing.T) {
	server := fixture.RunETCDServer(t)
	alice := screenplay.ActorNamed("Alice").WhoCan(use.ETCD().WithTheEndpoints(server.Endpoint()))
	require.NoError(t, alice.AttemptsTo(put.TheValue("postgresql://localhost:5432").ForTheKey("app/config/database")))

	t.Run("answers with the value of the key", func(t *testing.T) {
		theValue, err := alice.AsksFor(value.OfTheKey("app/config/database"))

		require.NoError(t, err)
		assert.Equal(t, "postgresql://localhost:5432", theValue)
	})

	t.Run("can be asserted on", func(t *testing.T) {
		require.NoError(t, alice.Should(
			see.The(value.OfTheKey("app/config/database")).Should(be.EqualTo("postgresql://localhost:5432")),
		))
	})

	t.Run("fails when the key does not exist", func(t *testing.T) {
		theValue, err := alice.AsksFor(value.OfTheKey("does/not/exist"))

		require.ErrorIs(t, err, etcd.ErrKeyNotFound)
		require.EqualError(t, err, `the key "does/not/exist" was not found`)
		assert.Empty(t, theValue)
	})

	t.Run("fails when the actor cannot use etcd", func(t *testing.T) {
		_, err := screenplay.ActorNamed("Bob").AsksFor(value.OfTheKey("app/config/database"))

		require.ErrorIs(t, err, screenplay.ErrActorMissingAbility)
	})

	t.Run("implements the stringer interface", func(t *testing.T) {
		assert.Equal(
			t,
			`value of the key "app/config/database"`,
			value.OfTheKey("app/config/database").String(),
		)
	})

	t.Run("implements the question interface", func(t *testing.T) {
		assert.Implements(t, (*screenplay.Question[string])(nil), value.OfTheKey("app/config/database"))
	})
}
