package put_test

import (
	"testing"

	"github.com/grandper/go-screenplay/v2/screenplay"
	screenplayfixture "github.com/grandper/go-screenplay/v2/screenplay/fixture"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grandper/go-screenplay-etcd/ability"
	"github.com/grandper/go-screenplay-etcd/action"
	"github.com/grandper/go-screenplay-etcd/action/put"
	"github.com/grandper/go-screenplay-etcd/fixture"
	etcdfixture "github.com/grandper/go-screenplay-etcd/internal/fixture"
)

func TestTheValues(t *testing.T) {
	server := fixture.RunETCDServer(t)
	client := server.NewClient(t)
	alice := screenplay.ActorNamed("Alice").WhoCan(ability.UseETCD().WithTheEndpoints(server.Endpoint()))
	config := map[string]string{
		"app/config/database": "postgresql://localhost:5432",
		"app/config/cache":    "redis://localhost:6379",
	}

	t.Run("puts several values", func(t *testing.T) {
		t.Run("in one transaction", func(t *testing.T) {
			revision := etcdfixture.RevisionOf(t, client)

			require.NoError(t, alice.AttemptsTo(put.TheValues(config)))

			assert.Equal(t, "postgresql://localhost:5432", etcdfixture.ValueOf(t, client, "app/config/database"))
			assert.Equal(t, "redis://localhost:6379", etcdfixture.ValueOf(t, client, "app/config/cache"))
			assert.Equal(t, revision+1, etcdfixture.RevisionOf(t, client), "both values are put in one revision")
		})

		t.Run("fails when no value is provided", func(t *testing.T) {
			require.ErrorIs(t, alice.AttemptsTo(put.TheValues(nil)), action.ErrKeyNotProvided)
			require.ErrorIs(t, alice.AttemptsTo(put.TheValues(map[string]string{})), action.ErrKeyNotProvided)
		})

		t.Run("fails when one of the keys is empty", func(t *testing.T) {
			err := alice.AttemptsTo(put.TheValues(map[string]string{"app/status": "ready", "": "lost"}))

			require.ErrorIs(t, err, action.ErrKeyNotProvided)
		})

		t.Run("fails when the actor cannot use etcd", func(t *testing.T) {
			err := screenplay.ActorNamed("Bob").AttemptsTo(put.TheValues(config))

			require.ErrorIs(t, err, screenplay.ErrActorMissingAbility)
		})
	})

	t.Run("is narrated with its values sorted by key", func(t *testing.T) {
		recorder := screenplayfixture.NewRecorder()
		adam := etcdfixture.NarratedActor("Adam", recorder, server.Endpoint())

		require.NoError(t, adam.AttemptsTo(put.TheValues(config)))

		assert.Equal(t,
			`Adam attempts to put the values "app/config/cache" = "redis://localhost:6379", `+
				`"app/config/database" = "postgresql://localhost:5432"`,
			recorder.Messages()[0],
		)
	})

	t.Run("can be kept secret", func(t *testing.T) {
		recorder := screenplayfixture.NewRecorder()
		adam := etcdfixture.NarratedActor("Adam", recorder, server.Endpoint())

		require.NoError(t, adam.AttemptsTo(put.TheValues(config).Secretly()))

		assert.Equal(t,
			`Adam attempts to put the values <secret> for the keys "app/config/cache", "app/config/database"`,
			recorder.Messages()[0],
		)
	})

	t.Run("implements the stringer interface", func(t *testing.T) {
		assert.Equal(
			t,
			`put the values "app/config/cache" = "redis://localhost:6379", "app/config/database" = "postgresql://localhost:5432"`,
			put.TheValues(config).String(),
		)
		assert.Equal(t,
			`put the values <secret> for the keys "app/config/cache", "app/config/database"`,
			put.TheValues(config).WhichShouldBeKeptSecret().String(),
		)
		assert.Equal(t, "put the values", put.TheValues(nil).String())
	})
}
