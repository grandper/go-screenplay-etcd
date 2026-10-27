package remove_test

import (
	"testing"

	"github.com/grandper/go-screenplay/v2/screenplay"
	screenplayfixture "github.com/grandper/go-screenplay/v2/screenplay/fixture"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grandper/go-screenplay-etcd/ability/use"
	"github.com/grandper/go-screenplay-etcd/action/put"
	"github.com/grandper/go-screenplay-etcd/action/remove"
	"github.com/grandper/go-screenplay-etcd/etcd"
	"github.com/grandper/go-screenplay-etcd/internal/fixture"
)

func TestTheKeysStartingWith(t *testing.T) {
	server := fixture.RunETCDServer(t)
	client := server.NewClient(t)
	alice := screenplay.ActorNamed("Alice").WhoCan(use.ETCD().WithTheEndpoints(server.Endpoint()))

	t.Run("removes the keys starting with a prefix", func(t *testing.T) {
		t.Run("leaving the other keys in place", func(t *testing.T) {
			require.NoError(t, alice.AttemptsTo(
				put.TheValue("redis://localhost:6379").ForTheKey("app/config/cache"),
				put.TheValue("postgresql://localhost:5432").ForTheKey("app/config/database"),
				put.TheValue("ready").ForTheKey("app/status"),
			))

			require.NoError(t, alice.AttemptsTo(remove.TheKeysStartingWith("app/config/")))

			fixture.AssertKeyDoesNotExist(t, client, "app/config/cache")
			fixture.AssertKeyDoesNotExist(t, client, "app/config/database")
			assert.Equal(t, "ready", fixture.ValueOf(t, client, "app/status"))
		})

		t.Run("when no key starts with the prefix, without an error", func(t *testing.T) {
			require.NoError(t, alice.AttemptsTo(remove.TheKeysStartingWith("nothing/")))
		})

		t.Run("fails when the prefix is not provided", func(t *testing.T) {
			err := alice.AttemptsTo(remove.TheKeysStartingWith(""))

			require.ErrorIs(t, err, etcd.ErrPrefixNotProvided)
			require.EqualError(t, err, "the prefix must be provided")
		})

		t.Run("fails when the actor cannot use etcd", func(t *testing.T) {
			err := screenplay.ActorNamed("Bob").AttemptsTo(remove.TheKeysStartingWith("app/config/"))

			require.ErrorIs(t, err, screenplay.ErrActorMissingAbility)
		})
	})

	t.Run("is narrated", func(t *testing.T) {
		recorder := screenplayfixture.NewRecorder()
		adam := fixture.NarratedActor("Adam", recorder, server.Endpoint())

		require.NoError(t, adam.AttemptsTo(remove.TheKeysStartingWith("app/config/")))

		assert.Equal(t, `Adam attempts to remove the keys starting with "app/config/"`, recorder.Messages()[0])
	})

	t.Run("implements the stringer interface", func(t *testing.T) {
		assert.Equal(
			t,
			`remove the keys starting with "app/config/"`,
			remove.TheKeysStartingWith("app/config/").String(),
		)
	})
}
