package status_test

import (
	"testing"

	"github.com/grandper/go-screenplay-etcd/question/status"

	"github.com/grandper/go-screenplay/v2/action/see"
	"github.com/grandper/go-screenplay/v2/resolution/be"
	"github.com/grandper/go-screenplay/v2/screenplay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grandper/go-screenplay-etcd/etcd"
	"github.com/grandper/go-screenplay-etcd/fixture"
	"github.com/grandper/go-screenplay-etcd/use"
)

func TestOfTheEndpoint(t *testing.T) {
	server := fixture.RunETCDServer(t)
	alice := screenplay.ActorNamed("Alice").WhoCan(use.ETCD().WithTheEndpoints(server.Endpoint()))

	t.Run("answers with the status of the endpoint", func(t *testing.T) {
		theStatus, err := alice.AsksFor(status.OfTheEndpoint(server.Endpoint()))

		require.NoError(t, err)
		assert.NotEmpty(t, theStatus.Version())
		assert.NotZero(t, theStatus.Leader())
		assert.Positive(t, theStatus.DBSize())
		assert.False(t, theStatus.IsLearner())
		assert.Empty(t, theStatus.Errors())
		assert.NotNil(t, theStatus.Errors())
	})

	t.Run("can be asserted on", func(t *testing.T) {
		require.NoError(t, alice.Should(
			see.The(status.OfTheEndpoint(server.Endpoint())).ShouldNotBe(be.Zero[*etcd.EndpointStatus]()),
		))
	})

	t.Run("fails when the endpoint cannot be reached", func(t *testing.T) {
		bob := screenplay.ActorNamed("Bob").WhoCan(
			use.ETCD().WithTheEndpoints("127.0.0.1:1").WithTimingOutAfter(500).Milliseconds(),
		)

		theStatus, err := bob.AsksFor(status.OfTheEndpoint("127.0.0.1:1"))

		require.Error(t, err)
		assert.Nil(t, theStatus)
	})

	t.Run("fails when the endpoint is not provided", func(t *testing.T) {
		_, err := alice.AsksFor(status.OfTheEndpoint(""))

		require.ErrorIs(t, err, etcd.ErrEndpointNotProvided)
	})

	t.Run("fails when the actor cannot use etcd", func(t *testing.T) {
		_, err := screenplay.ActorNamed("Carol").AsksFor(status.OfTheEndpoint(server.Endpoint()))

		require.ErrorIs(t, err, screenplay.ErrActorMissingAbility)
	})

	t.Run("implements the stringer interface", func(t *testing.T) {
		assert.Equal(
			t,
			`status of the endpoint "localhost:2379"`,
			status.OfTheEndpoint("localhost:2379").String(),
		)
	})

	t.Run("implements the question interface", func(t *testing.T) {
		assert.Implements(t,
			(*screenplay.Question[*etcd.EndpointStatus])(nil),
			status.OfTheEndpoint("localhost:2379"),
		)
	})
}
