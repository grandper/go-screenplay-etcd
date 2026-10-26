package remove_test

import (
	"testing"

	"github.com/grandper/go-screenplay/v2/screenplay"
	screenplayfixture "github.com/grandper/go-screenplay/v2/screenplay/fixture"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grandper/go-screenplay-etcd/action/put"
	"github.com/grandper/go-screenplay-etcd/action/remove"
	"github.com/grandper/go-screenplay-etcd/etcd"
	"github.com/grandper/go-screenplay-etcd/fixture"
	etcdfixture "github.com/grandper/go-screenplay-etcd/internal/fixture"
	"github.com/grandper/go-screenplay-etcd/use"
)

func TestTheKeys(t *testing.T) {
	server := fixture.RunETCDServer(t)
	client := server.NewClient(t)
	alice := screenplay.ActorNamed("Alice").WhoCan(use.ETCD().WithTheEndpoints(server.Endpoint()))

	t.Run("removes several keys", func(t *testing.T) {
		t.Run("in one transaction", func(t *testing.T) {
			require.NoError(t, alice.AttemptsTo(
				put.TheValue("redis://localhost:6379").ForTheKey("app/config/cache"),
				put.TheValue("postgresql://localhost:5432").ForTheKey("app/config/database"),
				put.TheValue("3").ForTheKey("app/config/retries"),
			))
			revision := etcdfixture.RevisionOf(t, client)

			require.NoError(t, alice.AttemptsTo(remove.TheKeys("app/config/cache", "app/config/database")))

			etcdfixture.AssertKeyDoesNotExist(t, client, "app/config/cache")
			etcdfixture.AssertKeyDoesNotExist(t, client, "app/config/database")
			assert.Equal(t, "3", etcdfixture.ValueOf(t, client, "app/config/retries"))
			assert.Equal(t, revision+1, etcdfixture.RevisionOf(t, client), "both keys are removed in one revision")
		})

		t.Run("some of which do not exist, without an error", func(t *testing.T) {
			require.NoError(t, alice.AttemptsTo(put.TheValue("3").ForTheKey("app/config/retries")))

			require.NoError(t, alice.AttemptsTo(remove.TheKeys("app/config/retries", "does/not/exist")))

			etcdfixture.AssertKeyDoesNotExist(t, client, "app/config/retries")
		})

		t.Run("fails when no key is provided", func(t *testing.T) {
			err := alice.AttemptsTo(remove.TheKeys())

			require.ErrorIs(t, err, etcd.ErrKeyNotProvided)
		})

		t.Run("fails when one of the keys is empty", func(t *testing.T) {
			err := alice.AttemptsTo(remove.TheKeys("app/config/cache", ""))

			require.ErrorIs(t, err, etcd.ErrKeyNotProvided)
		})

		t.Run("fails when the actor cannot use etcd", func(t *testing.T) {
			err := screenplay.ActorNamed("Bob").AttemptsTo(remove.TheKeys("app/config/cache", "app/config/database"))

			require.ErrorIs(t, err, screenplay.ErrActorMissingAbility)
		})
	})

	t.Run("is narrated", func(t *testing.T) {
		recorder := screenplayfixture.NewRecorder()
		adam := etcdfixture.NarratedActor("Adam", recorder, server.Endpoint())

		require.NoError(t, adam.AttemptsTo(remove.TheKeys("app/config/cache", "app/config/database")))

		assert.Equal(t,
			`Adam attempts to remove the keys "app/config/cache", "app/config/database"`,
			recorder.Messages()[0],
		)
	})

	t.Run("implements the stringer interface", func(t *testing.T) {
		assert.Equal(t,
			`remove the keys "app/config/cache", "app/config/database"`,
			remove.TheKeys("app/config/cache", "app/config/database").String(),
		)
		assert.Equal(t, "remove the keys", remove.TheKeys().String())
	})
}
