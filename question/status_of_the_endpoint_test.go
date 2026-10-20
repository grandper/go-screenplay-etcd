package question_test

import (
	"testing"

	"github.com/grandper/go-screenplay/v2/action/see"
	"github.com/grandper/go-screenplay/v2/resolution/be"
	"github.com/grandper/go-screenplay/v2/screenplay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grandper/go-screenplay-etcd/ability"
	"github.com/grandper/go-screenplay-etcd/fixture"
	"github.com/grandper/go-screenplay-etcd/question"
)

func TestStatusOfTheEndpoint(t *testing.T) {
	server := fixture.RunETCDServer(t)
	alice := screenplay.ActorNamed("Alice").WhoCan(ability.UseETCD().WithTheEndpoints(server.Endpoint()))

	t.Run("answers with the status of the endpoint", func(t *testing.T) {
		status, err := alice.AsksFor(question.StatusOfTheEndpoint(server.Endpoint()))

		require.NoError(t, err)
		assert.NotEmpty(t, status.Version())
		assert.NotZero(t, status.Leader())
		assert.Positive(t, status.DBSize())
		assert.False(t, status.IsLearner())
		assert.Empty(t, status.Errors())
		assert.NotNil(t, status.Errors())
	})

	t.Run("can be asserted on", func(t *testing.T) {
		require.NoError(t, alice.Should(
			see.The(question.StatusOfTheEndpoint(server.Endpoint())).ShouldNotBe(be.Zero[*ability.EndpointStatus]()),
		))
	})

	t.Run("fails when the endpoint cannot be reached", func(t *testing.T) {
		bob := screenplay.ActorNamed("Bob").WhoCan(
			ability.UseETCD().WithTheEndpoints("127.0.0.1:1").WithTimingOutAfter(500).Milliseconds(),
		)

		status, err := bob.AsksFor(question.StatusOfTheEndpoint("127.0.0.1:1"))

		require.Error(t, err)
		assert.Nil(t, status)
	})

	t.Run("fails when the endpoint is not provided", func(t *testing.T) {
		_, err := alice.AsksFor(question.StatusOfTheEndpoint(""))

		require.ErrorIs(t, err, question.ErrEndpointNotProvided)
	})

	t.Run("fails when the actor cannot use etcd", func(t *testing.T) {
		_, err := screenplay.ActorNamed("Carol").AsksFor(question.StatusOfTheEndpoint(server.Endpoint()))

		require.ErrorIs(t, err, screenplay.ErrActorMissingAbility)
	})

	t.Run("implements the stringer interface", func(t *testing.T) {
		assert.Equal(
			t,
			`status of the endpoint "localhost:2379"`,
			question.StatusOfTheEndpoint("localhost:2379").String(),
		)
	})

	t.Run("implements the question interface", func(t *testing.T) {
		assert.Implements(t,
			(*screenplay.Question[*ability.EndpointStatus])(nil),
			question.StatusOfTheEndpoint("localhost:2379"),
		)
	})
}
