package ability_test

import (
	"testing"

	"github.com/grandper/go-screenplay/v2/screenplay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grandper/go-screenplay-etcd/fixture"
	"github.com/grandper/go-screenplay-etcd/internal/ability"
)

func TestElections(t *testing.T) {
	server := fixture.RunETCDServer(t)
	newOperations := func(t *testing.T) *ability.Operations {
		t.Helper()
		theAbility := ability.UseETCD().WithTheEndpoints(server.Endpoint())
		t.Cleanup(func() { require.NoError(t, theAbility.Forget()) })

		return ability.TheAbilityOf(screenplay.ActorNamed("Alice").WhoCan(theAbility))
	}

	t.Run("campaigns in an election, leads it and resigns", func(t *testing.T) {
		alice, bob := newOperations(t), newOperations(t)

		leader, err := bob.ToRetrieveTheLeader(t.Context(), "primary-worker")
		require.NoError(t, err)
		assert.Empty(t, leader)

		require.NoError(t, alice.ToCampaign(t.Context(), "primary-worker", "node-1"))
		require.ErrorIs(t, alice.ToCampaign(t.Context(), "primary-worker", "node-1"), ability.ErrAlreadyCampaigning)

		leader, err = bob.ToRetrieveTheLeader(t.Context(), "primary-worker")
		require.NoError(t, err)
		assert.Equal(t, "node-1", leader)

		require.NoError(t, alice.ToResign(t.Context(), "primary-worker"))
		require.ErrorIs(t, alice.ToResign(t.Context(), "primary-worker"), ability.ErrNotCampaigning)

		leader, err = bob.ToRetrieveTheLeader(t.Context(), "primary-worker")
		require.NoError(t, err)
		assert.Empty(t, leader)
	})

	t.Run("resigns when the ability is forgotten", func(t *testing.T) {
		theAbility := ability.UseETCD().WithTheEndpoints(server.Endpoint())
		alice := ability.TheAbilityOf(screenplay.ActorNamed("Alice").WhoCan(theAbility))
		bob := newOperations(t)
		require.NoError(t, alice.ToCampaign(t.Context(), "forgotten", "node-1"))

		require.NoError(t, theAbility.Forget())

		require.NoError(t, bob.ToCampaign(t.Context(), "forgotten", "node-2"))
	})

	t.Run("fails every operation when the actor does not have the ability", func(t *testing.T) {
		noAbility := ability.TheAbilityOf(screenplay.ActorNamed("Bob"))

		require.ErrorIs(t, noAbility.ToCampaign(t.Context(), "primary-worker", "v"), screenplay.ErrActorMissingAbility)
		require.ErrorIs(t, noAbility.ToResign(t.Context(), "primary-worker"), screenplay.ErrActorMissingAbility)
		_, err := noAbility.ToRetrieveTheLeader(t.Context(), "primary-worker")
		require.ErrorIs(t, err, screenplay.ErrActorMissingAbility)
	})
}
