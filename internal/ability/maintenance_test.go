package ability_test

import (
	"testing"

	"github.com/grandper/go-screenplay/v2/screenplay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grandper/go-screenplay-etcd/fixture"
	"github.com/grandper/go-screenplay-etcd/internal/ability"
)

func TestMaintenance(t *testing.T) {
	server := fixture.RunETCDServer(t)
	theAbility := ability.TheAbilityOf(
		screenplay.ActorNamed("Alice").WhoCan(ability.UseETCD().WithTheEndpoints(server.Endpoint())),
	)

	t.Run("retrieves the status of an endpoint", func(t *testing.T) {
		status, err := theAbility.ToRetrieveTheStatusOf(t.Context(), server.Endpoint())

		require.NoError(t, err)
		assert.NotEmpty(t, status.Version())
		assert.NotZero(t, status.Leader())
		assert.Positive(t, status.DBSize())
		assert.False(t, status.IsLearner())
		assert.Empty(t, status.Errors())
	})

	t.Run("retrieves the active alarms", func(t *testing.T) {
		alarms, err := theAbility.ToRetrieveTheAlarms(t.Context())

		require.NoError(t, err)
		assert.NotNil(t, alarms)
		assert.Empty(t, alarms)
	})

	t.Run("fails every operation when the actor does not have the ability", func(t *testing.T) {
		noAbility := ability.TheAbilityOf(screenplay.ActorNamed("Bob"))

		_, err := noAbility.ToRetrieveTheStatusOf(t.Context(), server.Endpoint())
		require.ErrorIs(t, err, screenplay.ErrActorMissingAbility)
		_, err = noAbility.ToRetrieveTheAlarms(t.Context())
		require.ErrorIs(t, err, screenplay.ErrActorMissingAbility)
	})
}
