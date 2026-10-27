package campaign_test

import (
	"testing"

	"github.com/grandper/go-screenplay/v2/screenplay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grandper/go-screenplay-etcd/ability/use"
	"github.com/grandper/go-screenplay-etcd/etcd"
	"github.com/grandper/go-screenplay-etcd/internal/action/campaign"
	"github.com/grandper/go-screenplay-etcd/internal/fixture"
	"github.com/grandper/go-screenplay-etcd/question/leader"
)

func TestInTheElection(t *testing.T) {
	server := fixture.RunETCDServer(t)
	newActor := func(t *testing.T, name string) *screenplay.Actor {
		t.Helper()
		actor := screenplay.ActorNamed(name).WhoCan(use.ETCD().WithTheEndpoints(server.Endpoint()))
		t.Cleanup(func() { require.NoError(t, actor.Exit()) })

		return actor
	}

	t.Run("announces the leader under the name of the actor", func(t *testing.T) {
		alice := newActor(t, "Alice")

		require.NoError(t, alice.AttemptsTo(campaign.InTheElection("primary-worker")))

		theLeader, err := alice.AsksFor(leader.OfTheElection("primary-worker"))
		require.NoError(t, err)
		assert.Equal(t, "Alice", theLeader)
	})

	t.Run("announces the leader under the value it is given", func(t *testing.T) {
		alice := newActor(t, "Alice")

		require.NoError(t, alice.AttemptsTo(campaign.InTheElection("backup-worker").WithTheValue("node-1")))

		theLeader, err := alice.AsksFor(leader.OfTheElection("backup-worker"))
		require.NoError(t, err)
		assert.Equal(t, "node-1", theLeader)
	})

	t.Run("announces the leader under an empty value when that is what it is given", func(t *testing.T) {
		alice := newActor(t, "Alice")

		require.NoError(t, alice.AttemptsTo(campaign.InTheElection("silent-worker").WithTheValue("")))

		theLeader, err := alice.AsksFor(leader.OfTheElection("silent-worker"))
		require.NoError(t, err)
		assert.Empty(t, theLeader, "the given value wins over the name of the actor")
	})

	t.Run("fails when the actor already campaigns in the election", func(t *testing.T) {
		alice := newActor(t, "Alice")
		require.NoError(t, alice.AttemptsTo(campaign.InTheElection("primary-worker")))

		err := alice.AttemptsTo(campaign.InTheElection("primary-worker"))

		require.ErrorIs(t, err, etcd.ErrAlreadyCampaigning)
		require.EqualError(t, err, `the actor already campaigns in the election "primary-worker"`)
	})

	t.Run("reports a missing name before reaching the store", func(t *testing.T) {
		err := screenplay.ActorNamed("Carol").AttemptsTo(campaign.InTheElection(""))

		require.ErrorIs(t, err, etcd.ErrNameNotProvided)
		require.NotErrorIs(t, err, screenplay.ErrActorMissingAbility)
	})

	t.Run("fails when the actor cannot use etcd", func(t *testing.T) {
		err := screenplay.ActorNamed("Carol").AttemptsTo(campaign.InTheElection("primary-worker"))

		require.ErrorIs(t, err, screenplay.ErrActorMissingAbility)
	})

	t.Run("implements the stringer interface", func(t *testing.T) {
		assert.Equal(t, `campaign in the election "primary-worker"`, campaign.InTheElection("primary-worker").String())
		assert.Equal(t,
			`campaign in the election "primary-worker" with the value "node-1"`,
			campaign.InTheElection("primary-worker").WithTheValue("node-1").String(),
		)
		assert.Equal(t,
			`campaign in the election "primary-worker" with the value ""`,
			campaign.InTheElection("primary-worker").WithTheValue("").String(),
		)
	})

	t.Run("implements the performable interface", func(t *testing.T) {
		assert.Implements(t, (*screenplay.Performable)(nil), campaign.InTheElection("primary-worker"))
	})
}
