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

func TestRevisionOfTheKey(t *testing.T) {
	server := fixture.RunETCDServer(t)
	alice := screenplay.ActorNamed("Alice").WhoCan(ability.UseETCD().WithTheEndpoints(server.Endpoint()))

	t.Run("answers with the revision at which the key was last modified", func(t *testing.T) {
		require.NoError(t, alice.AttemptsTo(put.TheValue("1").ForTheKey("app/config/retries")))
		before, err := alice.AsksFor(question.RevisionOfTheKey("app/config/retries"))
		require.NoError(t, err)
		assert.Positive(t, before)

		require.NoError(t, alice.AttemptsTo(put.TheValue("2").ForTheKey("app/config/retries")))

		after, err := alice.AsksFor(question.RevisionOfTheKey("app/config/retries"))
		require.NoError(t, err)
		assert.Equal(t, before+1, after)
		require.NoError(t, alice.Should(
			see.The(question.RevisionOfTheKey("app/config/retries")).Should(be.GreaterThan(before)),
		))
	})

	t.Run("fails when the key does not exist", func(t *testing.T) {
		revision, err := alice.AsksFor(question.RevisionOfTheKey("does/not/exist"))

		require.ErrorIs(t, err, question.ErrKeyNotFound)
		require.EqualError(t, err, `the key "does/not/exist" was not found`)
		assert.Zero(t, revision)
	})

	t.Run("fails when the key is not provided", func(t *testing.T) {
		_, err := alice.AsksFor(question.RevisionOfTheKey(""))

		require.ErrorIs(t, err, question.ErrKeyNotProvided)
	})

	t.Run("fails when the actor cannot use etcd", func(t *testing.T) {
		_, err := screenplay.ActorNamed("Bob").AsksFor(question.RevisionOfTheKey("app/config/retries"))

		require.ErrorIs(t, err, screenplay.ErrActorMissingAbility)
	})

	t.Run("implements the stringer interface", func(t *testing.T) {
		assert.Equal(
			t,
			`revision of the key "app/config/retries"`,
			question.RevisionOfTheKey("app/config/retries").String(),
		)
	})

	t.Run("implements the question interface", func(t *testing.T) {
		assert.Implements(t, (*screenplay.Question[int64])(nil), question.RevisionOfTheKey("app/config/retries"))
	})
}
