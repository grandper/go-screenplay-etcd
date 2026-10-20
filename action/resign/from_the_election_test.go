package resign_test

import (
	"testing"

	"github.com/grandper/go-screenplay/v2/action/see"
	"github.com/grandper/go-screenplay/v2/resolution/be"
	"github.com/grandper/go-screenplay/v2/screenplay"
	screenplayfixture "github.com/grandper/go-screenplay/v2/screenplay/fixture"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grandper/go-screenplay-etcd/ability"
	"github.com/grandper/go-screenplay-etcd/action"
	"github.com/grandper/go-screenplay-etcd/action/campaign"
	"github.com/grandper/go-screenplay-etcd/action/resign"
	"github.com/grandper/go-screenplay-etcd/fixture"
	etcdfixture "github.com/grandper/go-screenplay-etcd/internal/fixture"
	"github.com/grandper/go-screenplay-etcd/question"
)

func TestFromTheElection(t *testing.T) {
	server := fixture.RunETCDServer(t)
	newActor := func(t *testing.T, name string) *screenplay.Actor {
		t.Helper()
		actor := screenplay.ActorNamed(name).WhoCan(ability.UseETCD().WithTheEndpoints(server.Endpoint()))
		t.Cleanup(func() { require.NoError(t, actor.Exit()) })

		return actor
	}

	t.Run("resigns from an election", func(t *testing.T) {
		t.Run("leaving it without a leader", func(t *testing.T) {
			alice := newActor(t, "Alice")
			require.NoError(t, alice.AttemptsTo(campaign.InTheElection("primary-worker")))

			require.NoError(t, alice.AttemptsTo(resign.FromTheElection("primary-worker")))

			require.NoError(
				t,
				alice.Should(see.The(question.LeaderOfTheElection("primary-worker")).Should(be.Empty[string]())),
			)
		})

		t.Run("and can campaign again", func(t *testing.T) {
			alice := newActor(t, "Alice")

			require.NoError(t, alice.AttemptsTo(
				campaign.InTheElection("primary-worker"),
				resign.FromTheElection("primary-worker"),
				campaign.InTheElection("primary-worker"),
			))
		})

		t.Run("fails when the actor does not campaign in it", func(t *testing.T) {
			alice := newActor(t, "Alice")

			err := alice.AttemptsTo(resign.FromTheElection("primary-worker"))

			require.ErrorIs(t, err, ability.ErrNotCampaigning)
			require.EqualError(t, err, `the actor does not campaign in the election "primary-worker"`)
		})

		t.Run("fails when the name is not provided", func(t *testing.T) {
			alice := newActor(t, "Alice")

			require.ErrorIs(t, alice.AttemptsTo(resign.FromTheElection("")), action.ErrNameNotProvided)
		})

		t.Run("fails when the actor cannot use etcd", func(t *testing.T) {
			err := screenplay.ActorNamed("Carol").AttemptsTo(resign.FromTheElection("primary-worker"))

			require.ErrorIs(t, err, screenplay.ErrActorMissingAbility)
		})
	})

	t.Run("is narrated", func(t *testing.T) {
		recorder := screenplayfixture.NewRecorder()
		adam := etcdfixture.NarratedActor("Adam", recorder, server.Endpoint())
		t.Cleanup(func() { require.NoError(t, adam.Exit()) })

		require.NoError(
			t,
			adam.AttemptsTo(campaign.InTheElection("primary-worker"), resign.FromTheElection("primary-worker")),
		)

		assert.Equal(t, `Adam attempts to resign from the election "primary-worker"`, recorder.Messages()[2])
	})

	t.Run("implements the stringer interface", func(t *testing.T) {
		assert.Equal(t, `resign from the election "primary-worker"`, resign.FromTheElection("primary-worker").String())
	})
}
