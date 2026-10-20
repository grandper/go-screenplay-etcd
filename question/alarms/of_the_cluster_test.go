package alarms_test

import (
	"testing"

	"github.com/grandper/go-screenplay/v2/action/see"
	"github.com/grandper/go-screenplay/v2/resolution/be"
	"github.com/grandper/go-screenplay/v2/screenplay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grandper/go-screenplay-etcd/ability"
	"github.com/grandper/go-screenplay-etcd/fixture"
	"github.com/grandper/go-screenplay-etcd/question/alarms"
)

func TestOfTheCluster(t *testing.T) {
	server := fixture.RunETCDServer(t)
	alice := screenplay.ActorNamed("Alice").WhoCan(ability.UseETCD().WithTheEndpoints(server.Endpoint()))

	t.Run("answers with no alarm on a healthy cluster", func(t *testing.T) {
		theAlarms, err := alice.AsksFor(alarms.OfTheCluster())

		require.NoError(t, err)
		assert.NotNil(t, theAlarms)
		assert.Empty(t, theAlarms)
	})

	t.Run("can be asserted on", func(t *testing.T) {
		require.NoError(t, alice.Should(see.The(alarms.OfTheCluster()).Should(be.Empty[[]*ability.Alarm]())))
	})

	t.Run("fails when the actor cannot use etcd", func(t *testing.T) {
		_, err := screenplay.ActorNamed("Carol").AsksFor(alarms.OfTheCluster())

		require.ErrorIs(t, err, screenplay.ErrActorMissingAbility)
	})

	t.Run("implements the stringer interface", func(t *testing.T) {
		assert.Equal(t, "alarms of the cluster", alarms.OfTheCluster().String())
	})

	t.Run("implements the question interface", func(t *testing.T) {
		assert.Implements(t, (*screenplay.Question[[]*ability.Alarm])(nil), alarms.OfTheCluster())
	})
}
