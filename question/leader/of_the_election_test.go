package leader_test

import (
	"testing"

	"github.com/grandper/go-screenplay/v2/action/see"
	"github.com/grandper/go-screenplay/v2/resolution/be"
	"github.com/grandper/go-screenplay/v2/screenplay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grandper/go-screenplay-etcd/ability/use"
	"github.com/grandper/go-screenplay-etcd/action/campaign"
	"github.com/grandper/go-screenplay-etcd/etcd"
	"github.com/grandper/go-screenplay-etcd/internal/fixture"
	"github.com/grandper/go-screenplay-etcd/question/leader"
)

func TestOfTheElection(t *testing.T) {
	server := fixture.RunETCDServer(t)
	alice := screenplay.ActorNamed("Alice").WhoCan(use.ETCD().WithTheEndpoints(server.Endpoint()))
	t.Cleanup(func() { require.NoError(t, alice.Exit()) })
	bob := screenplay.ActorNamed("Bob").WhoCan(use.ETCD().WithTheEndpoints(server.Endpoint()))
	t.Cleanup(func() { require.NoError(t, bob.Exit()) })

	t.Run("answers with the value of the leader, whoever asks", func(t *testing.T) {
		require.NoError(t, alice.AttemptsTo(campaign.InTheElection("primary-worker").WithTheValue("node-1")))

		theLeader, err := bob.AsksFor(leader.OfTheElection("primary-worker"))

		require.NoError(t, err)
		assert.Equal(t, "node-1", theLeader)
		require.NoError(
			t,
			bob.Should(see.The(leader.OfTheElection("primary-worker")).Should(be.EqualTo("node-1"))),
		)
	})

	t.Run("answers with an empty value when the election has no leader", func(t *testing.T) {
		theLeader, err := bob.AsksFor(leader.OfTheElection("nobody-campaigns"))

		require.NoError(t, err)
		assert.Empty(t, theLeader)
	})

	t.Run("fails when the name is not provided", func(t *testing.T) {
		_, err := alice.AsksFor(leader.OfTheElection(""))

		require.ErrorIs(t, err, etcd.ErrNameNotProvided)
	})

	t.Run("fails when the actor cannot use etcd", func(t *testing.T) {
		_, err := screenplay.ActorNamed("Carol").AsksFor(leader.OfTheElection("primary-worker"))

		require.ErrorIs(t, err, screenplay.ErrActorMissingAbility)
	})

	t.Run("implements the stringer interface", func(t *testing.T) {
		assert.Equal(
			t,
			`leader of the election "primary-worker"`,
			leader.OfTheElection("primary-worker").String(),
		)
	})

	t.Run("implements the question interface", func(t *testing.T) {
		assert.Implements(t, (*screenplay.Question[string])(nil), leader.OfTheElection("primary-worker"))
	})
}
