package question_test

import (
	"testing"

	"github.com/grandper/go-screenplay/v2/action/see"
	"github.com/grandper/go-screenplay/v2/resolution/be"
	"github.com/grandper/go-screenplay/v2/screenplay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grandper/go-screenplay-etcd/ability"
	"github.com/grandper/go-screenplay-etcd/action/campaign"
	"github.com/grandper/go-screenplay-etcd/fixture"
	"github.com/grandper/go-screenplay-etcd/question"
)

func TestLeaderOfTheElection(t *testing.T) {
	server := fixture.RunETCDServer(t)
	alice := screenplay.ActorNamed("Alice").WhoCan(ability.UseETCD().WithTheEndpoints(server.Endpoint()))
	t.Cleanup(func() { require.NoError(t, alice.Exit()) })
	bob := screenplay.ActorNamed("Bob").WhoCan(ability.UseETCD().WithTheEndpoints(server.Endpoint()))
	t.Cleanup(func() { require.NoError(t, bob.Exit()) })

	t.Run("answers with the value of the leader, whoever asks", func(t *testing.T) {
		require.NoError(t, alice.AttemptsTo(campaign.InTheElection("primary-worker").WithTheValue("node-1")))

		leader, err := bob.AsksFor(question.LeaderOfTheElection("primary-worker"))

		require.NoError(t, err)
		assert.Equal(t, "node-1", leader)
		require.NoError(
			t,
			bob.Should(see.The(question.LeaderOfTheElection("primary-worker")).Should(be.EqualTo("node-1"))),
		)
	})

	t.Run("answers with an empty value when the election has no leader", func(t *testing.T) {
		leader, err := bob.AsksFor(question.LeaderOfTheElection("nobody-campaigns"))

		require.NoError(t, err)
		assert.Empty(t, leader)
	})

	t.Run("fails when the name is not provided", func(t *testing.T) {
		_, err := alice.AsksFor(question.LeaderOfTheElection(""))

		require.ErrorIs(t, err, question.ErrNameNotProvided)
	})

	t.Run("fails when the actor cannot use etcd", func(t *testing.T) {
		_, err := screenplay.ActorNamed("Carol").AsksFor(question.LeaderOfTheElection("primary-worker"))

		require.ErrorIs(t, err, screenplay.ErrActorMissingAbility)
	})

	t.Run("implements the stringer interface", func(t *testing.T) {
		assert.Equal(
			t,
			`leader of the election "primary-worker"`,
			question.LeaderOfTheElection("primary-worker").String(),
		)
	})

	t.Run("implements the question interface", func(t *testing.T) {
		assert.Implements(t, (*screenplay.Question[string])(nil), question.LeaderOfTheElection("primary-worker"))
	})
}
