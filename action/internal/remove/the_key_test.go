package remove_test

import (
	"fmt"
	"testing"

	"github.com/grandper/go-screenplay/v2/screenplay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grandper/go-screenplay-etcd/action/internal/remove"
	"github.com/grandper/go-screenplay-etcd/action/put"
	"github.com/grandper/go-screenplay-etcd/etcd"
	"github.com/grandper/go-screenplay-etcd/fixture"
	etcdfixture "github.com/grandper/go-screenplay-etcd/internal/fixture"
	"github.com/grandper/go-screenplay-etcd/question/revision"
	"github.com/grandper/go-screenplay-etcd/use"
)

func TestTheKey(t *testing.T) {
	server := fixture.RunETCDServer(t)
	client := server.NewClient(t)
	alice := screenplay.ActorNamed("Alice").WhoCan(use.ETCD().WithTheEndpoints(server.Endpoint()))

	t.Run("removes the key", func(t *testing.T) {
		require.NoError(t, alice.AttemptsTo(put.TheValue("redis://localhost:6379").ForTheKey("app/config/cache")))

		require.NoError(t, alice.AttemptsTo(remove.TheKey("app/config/cache")))

		etcdfixture.AssertKeyDoesNotExist(t, client, "app/config/cache")
	})

	t.Run("removes a key that does not exist without an error", func(t *testing.T) {
		require.NoError(t, alice.AttemptsTo(remove.TheKey("does/not/exist")))
	})

	t.Run("removes the key only if it is unchanged since the revision", func(t *testing.T) {
		require.NoError(t, alice.AttemptsTo(put.TheValue("1").ForTheKey("app/config/counter")))
		theRevision, err := alice.AsksFor(revision.OfTheKey("app/config/counter"))
		require.NoError(t, err)
		require.NoError(t, alice.AttemptsTo(put.TheValue("2").ForTheKey("app/config/counter")))

		err = alice.AttemptsTo(remove.TheKey("app/config/counter").IfUnchangedSinceTheRevision(theRevision))

		require.ErrorIs(t, err, etcd.ErrKeyModified)
		require.EqualError(
			t,
			err,
			fmt.Sprintf(`the key "app/config/counter" was modified since the revision %d`, theRevision),
		)
		assert.Equal(t, "2", etcdfixture.ValueOf(t, client, "app/config/counter"), "the store is left untouched")

		theRevision, err = alice.AsksFor(revision.OfTheKey("app/config/counter"))
		require.NoError(t, err)
		require.NoError(
			t,
			alice.AttemptsTo(remove.TheKey("app/config/counter").IfUnchangedSinceTheRevision(theRevision)),
		)
		etcdfixture.AssertKeyDoesNotExist(t, client, "app/config/counter")
	})

	t.Run("reports a missing key before reaching the store", func(t *testing.T) {
		err := screenplay.ActorNamed("Carol").AttemptsTo(remove.TheKey(""))

		require.ErrorIs(t, err, etcd.ErrKeyNotProvided)
		require.NotErrorIs(t, err, screenplay.ErrActorMissingAbility)
	})

	t.Run("fails when the actor cannot use etcd", func(t *testing.T) {
		err := screenplay.ActorNamed("Carol").AttemptsTo(remove.TheKey("app/config/cache"))

		require.ErrorIs(t, err, screenplay.ErrActorMissingAbility)
	})

	t.Run("implements the stringer interface", func(t *testing.T) {
		assert.Equal(t, `remove the key "app/config/cache"`, remove.TheKey("app/config/cache").String())
		assert.Equal(
			t,
			`remove the key "app/config/cache" if unchanged since the revision 42`,
			remove.TheKey("app/config/cache").IfUnchangedSinceTheRevision(42).String(),
		)
	})

	t.Run("implements the performable interface", func(t *testing.T) {
		assert.Implements(t, (*screenplay.Performable)(nil), remove.TheKey("app/config/cache"))
	})
}
