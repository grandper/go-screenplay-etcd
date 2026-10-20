package remove_test

import (
	"fmt"
	"testing"

	"github.com/grandper/go-screenplay-etcd/question/revision"

	"github.com/grandper/go-screenplay/v2/screenplay"
	screenplayfixture "github.com/grandper/go-screenplay/v2/screenplay/fixture"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grandper/go-screenplay-etcd/ability"
	"github.com/grandper/go-screenplay-etcd/action"
	"github.com/grandper/go-screenplay-etcd/action/put"
	"github.com/grandper/go-screenplay-etcd/action/remove"
	"github.com/grandper/go-screenplay-etcd/fixture"
	etcdfixture "github.com/grandper/go-screenplay-etcd/internal/fixture"
)

func TestTheKey(t *testing.T) {
	server := fixture.RunETCDServer(t)
	client := server.NewClient(t)
	alice := screenplay.ActorNamed("Alice").WhoCan(ability.UseETCD().WithTheEndpoints(server.Endpoint()))

	t.Run("removes a key", func(t *testing.T) {
		t.Run("successfully", func(t *testing.T) {
			require.NoError(t, alice.AttemptsTo(put.TheValue("redis://localhost:6379").ForTheKey("app/config/cache")))

			require.NoError(t, alice.AttemptsTo(remove.TheKey("app/config/cache")))

			etcdfixture.AssertKeyDoesNotExist(t, client, "app/config/cache")
		})

		t.Run("leaving the other keys in place", func(t *testing.T) {
			require.NoError(t, alice.AttemptsTo(
				put.TheValue("redis://localhost:6379").ForTheKey("app/config/cache"),
				put.TheValue("postgresql://localhost:5432").ForTheKey("app/config/database"),
			))

			require.NoError(t, alice.AttemptsTo(remove.TheKey("app/config/cache")))

			assert.Equal(t, "postgresql://localhost:5432", etcdfixture.ValueOf(t, client, "app/config/database"))
		})

		t.Run("that does not exist, without an error", func(t *testing.T) {
			require.NoError(t, alice.AttemptsTo(remove.TheKey("does/not/exist")))
		})

		t.Run("only if unchanged since a revision", func(t *testing.T) {
			require.NoError(t, alice.AttemptsTo(put.TheValue("1").ForTheKey("app/config/counter")))
			theRevision, err := alice.AsksFor(revision.OfTheKey("app/config/counter"))
			require.NoError(t, err)
			require.NoError(t, alice.AttemptsTo(put.TheValue("2").ForTheKey("app/config/counter")))

			err = alice.AttemptsTo(remove.TheKey("app/config/counter").IfUnchangedSinceTheRevision(theRevision))
			require.ErrorIs(t, err, action.ErrKeyModified)
			require.EqualError(t, err,
				fmt.Sprintf(`the key "app/config/counter" was modified since the revision %d`, theRevision),
			)
			assert.Equal(t, "2", etcdfixture.ValueOf(t, client, "app/config/counter"))

			require.NoError(
				t,
				alice.AttemptsTo(remove.TheKey("app/config/counter").IfUnchangedSinceTheRevision(theRevision+1)),
			)
			etcdfixture.AssertKeyDoesNotExist(t, client, "app/config/counter")
		})

		t.Run("fails when the key is not provided", func(t *testing.T) {
			err := alice.AttemptsTo(remove.TheKey(""))

			require.ErrorIs(t, err, action.ErrKeyNotProvided)
			require.EqualError(t, err, "the key must be provided")
		})

		t.Run("fails when the actor cannot use etcd", func(t *testing.T) {
			err := screenplay.ActorNamed("Bob").AttemptsTo(remove.TheKey("app/config/cache"))

			require.ErrorIs(t, err, screenplay.ErrActorMissingAbility)
		})
	})

	t.Run("is narrated", func(t *testing.T) {
		recorder := screenplayfixture.NewRecorder()
		adam := etcdfixture.NarratedActor("Adam", recorder, server.Endpoint())

		require.NoError(t, adam.AttemptsTo(remove.TheKey("app/config/cache")))

		assert.Equal(t, `Adam attempts to remove the key "app/config/cache"`, recorder.Messages()[0])
	})

	t.Run("implements the stringer interface", func(t *testing.T) {
		assert.Equal(t, `remove the key "app/config/cache"`, remove.TheKey("app/config/cache").String())
		assert.Equal(t,
			`remove the key "app/config/cache" if unchanged since the revision 42`,
			remove.TheKey("app/config/cache").IfUnchangedSinceTheRevision(42).String(),
		)
	})
}
