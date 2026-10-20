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

func TestValueOfTheKey(t *testing.T) {
	server := fixture.RunETCDServer(t)
	alice := screenplay.ActorNamed("Alice").WhoCan(ability.UseETCD().WithTheEndpoints(server.Endpoint()))
	require.NoError(t, alice.AttemptsTo(put.TheValue("postgresql://localhost:5432").ForTheKey("app/config/database")))

	t.Run("answers with the value of the key", func(t *testing.T) {
		value, err := alice.AsksFor(question.ValueOfTheKey("app/config/database"))

		require.NoError(t, err)
		assert.Equal(t, "postgresql://localhost:5432", value)
	})

	t.Run("can be asserted on", func(t *testing.T) {
		require.NoError(t, alice.Should(
			see.The(question.ValueOfTheKey("app/config/database")).Should(be.EqualTo("postgresql://localhost:5432")),
		))
	})

	t.Run("fails when the key does not exist", func(t *testing.T) {
		value, err := alice.AsksFor(question.ValueOfTheKey("does/not/exist"))

		require.ErrorIs(t, err, question.ErrKeyNotFound)
		require.EqualError(t, err, `the key "does/not/exist" was not found`)
		assert.Empty(t, value)
	})

	t.Run("fails when the actor cannot use etcd", func(t *testing.T) {
		_, err := screenplay.ActorNamed("Bob").AsksFor(question.ValueOfTheKey("app/config/database"))

		require.ErrorIs(t, err, screenplay.ErrActorMissingAbility)
	})

	t.Run("implements the stringer interface", func(t *testing.T) {
		assert.Equal(
			t,
			`value of the key "app/config/database"`,
			question.ValueOfTheKey("app/config/database").String(),
		)
	})

	t.Run("implements the question interface", func(t *testing.T) {
		assert.Implements(t, (*screenplay.Question[string])(nil), question.ValueOfTheKey("app/config/database"))
	})
}
