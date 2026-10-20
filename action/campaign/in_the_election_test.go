package campaign_test

import (
	"testing"
	"time"

	"github.com/grandper/go-screenplay-etcd/question/leader"

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
)

func TestInTheElection(t *testing.T) {
	server := fixture.RunETCDServer(t)
	newActor := func(t *testing.T, name string) *screenplay.Actor {
		t.Helper()
		actor := screenplay.ActorNamed(name).WhoCan(ability.UseETCD().WithTheEndpoints(server.Endpoint()))
		t.Cleanup(func() { require.NoError(t, actor.Exit()) })

		return actor
	}

	t.Run("campaigns in an election", func(t *testing.T) {
		t.Run("and becomes its leader, under the name of the actor", func(t *testing.T) {
			alice := newActor(t, "Alice")

			require.NoError(t, alice.AttemptsTo(campaign.InTheElection("primary-worker")))

			require.NoError(
				t,
				alice.Should(see.The(leader.OfTheElection("primary-worker")).Should(be.EqualTo("Alice"))),
			)
		})

		t.Run("with a value of its own", func(t *testing.T) {
			alice := newActor(t, "Alice")

			require.NoError(t, alice.AttemptsTo(campaign.InTheElection("backup-worker").WithTheValue("node-1")))

			require.NoError(
				t,
				alice.Should(see.The(leader.OfTheElection("backup-worker")).Should(be.EqualTo("node-1"))),
			)
		})

		t.Run("waiting for the leader to resign", func(t *testing.T) {
			alice, bob := newActor(t, "Alice"), newActor(t, "Bob")
			require.NoError(t, alice.AttemptsTo(campaign.InTheElection("primary-worker")))

			elected := make(chan error, 1)
			go func() { elected <- bob.AttemptsTo(campaign.InTheElection("primary-worker")) }()

			select {
			case err := <-elected:
				t.Fatalf("Bob was elected while Alice leads: %v", err)
			case <-time.After(500 * time.Millisecond):
			}

			require.NoError(t, alice.AttemptsTo(resign.FromTheElection("primary-worker")))
			select {
			case err := <-elected:
				require.NoError(t, err)
			case <-time.After(15 * time.Second):
				t.Fatal("Bob was not elected after Alice resigned")
			}
			require.NoError(
				t,
				bob.Should(see.The(leader.OfTheElection("primary-worker")).Should(be.EqualTo("Bob"))),
			)
		})

		t.Run("fails when the actor already campaigns in it", func(t *testing.T) {
			alice := newActor(t, "Alice")
			require.NoError(t, alice.AttemptsTo(campaign.InTheElection("primary-worker")))

			err := alice.AttemptsTo(campaign.InTheElection("primary-worker"))

			require.ErrorIs(t, err, ability.ErrAlreadyCampaigning)
			require.EqualError(t, err, `the actor already campaigns in the election "primary-worker"`)
		})

		t.Run("fails when the name is not provided", func(t *testing.T) {
			alice := newActor(t, "Alice")

			require.ErrorIs(t, alice.AttemptsTo(campaign.InTheElection("")), action.ErrNameNotProvided)
		})

		t.Run("fails when the actor cannot use etcd", func(t *testing.T) {
			err := screenplay.ActorNamed("Carol").AttemptsTo(campaign.InTheElection("primary-worker"))

			require.ErrorIs(t, err, screenplay.ErrActorMissingAbility)
		})
	})

	t.Run("is narrated", func(t *testing.T) {
		recorder := screenplayfixture.NewRecorder()
		adam := etcdfixture.NarratedActor("Adam", recorder, server.Endpoint())
		t.Cleanup(func() { require.NoError(t, adam.Exit()) })

		require.NoError(t, adam.AttemptsTo(campaign.InTheElection("primary-worker").WithTheValue("node-1")))

		assert.Equal(
			t,
			`Adam attempts to campaign in the election "primary-worker" with the value "node-1"`,
			recorder.Messages()[0],
		)
	})

	t.Run("implements the stringer interface", func(t *testing.T) {
		assert.Equal(t, `campaign in the election "primary-worker"`, campaign.InTheElection("primary-worker").String())
		assert.Equal(t,
			`campaign in the election "primary-worker" with the value "node-1"`,
			campaign.InTheElection("primary-worker").WithTheValue("node-1").String(),
		)
	})
}
