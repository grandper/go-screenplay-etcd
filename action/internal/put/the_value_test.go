package put_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/grandper/go-screenplay/v2/screenplay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grandper/go-screenplay-etcd/action/internal/put"
	"github.com/grandper/go-screenplay-etcd/action/remove"
	"github.com/grandper/go-screenplay-etcd/etcd"
	"github.com/grandper/go-screenplay-etcd/fixture"
	etcdfixture "github.com/grandper/go-screenplay-etcd/internal/fixture"
	"github.com/grandper/go-screenplay-etcd/question/key"
	"github.com/grandper/go-screenplay-etcd/question/revision"
	"github.com/grandper/go-screenplay-etcd/use"
)

func TestTheValue(t *testing.T) {
	server := fixture.RunETCDServer(t)
	client := server.NewClient(t)
	alice := screenplay.ActorNamed("Alice").WhoCan(use.ETCD().WithTheEndpoints(server.Endpoint()))
	t.Cleanup(func() { require.NoError(t, alice.Exit()) })

	t.Run("puts the value for the key, which never expires", func(t *testing.T) {
		require.NoError(
			t,
			alice.AttemptsTo(put.TheValue("postgresql://localhost:5432").ForTheKey("app/config/database")),
		)

		theKey, err := alice.AsksFor(key.Named("app/config/database"))
		require.NoError(t, err)
		assert.Equal(t, "postgresql://localhost:5432", theKey.Value())
		assert.Zero(t, theKey.Lease())
	})

	t.Run("puts the value only if the key is absent", func(t *testing.T) {
		require.NoError(t, alice.AttemptsTo(remove.TheKey("app/config/flag")))
		require.NoError(t, alice.AttemptsTo(put.TheValue("initial").ForTheKey("app/config/flag").IfTheKeyIsAbsent()))

		err := alice.AttemptsTo(put.TheValue("other").ForTheKey("app/config/flag").IfTheKeyIsAbsent())

		require.ErrorIs(t, err, etcd.ErrKeyAlreadyExists)
		require.EqualError(t, err, `the key "app/config/flag" already exists`)
		assert.Equal(t, "initial", etcdfixture.ValueOf(t, client, "app/config/flag"), "the store is left untouched")
	})

	t.Run("puts the value only if the key is unchanged since the revision", func(t *testing.T) {
		require.NoError(t, alice.AttemptsTo(put.TheValue("1").ForTheKey("app/config/counter")))
		theRevision, err := alice.AsksFor(revision.OfTheKey("app/config/counter"))
		require.NoError(t, err)

		require.NoError(t, alice.AttemptsTo(
			put.TheValue("2").ForTheKey("app/config/counter").IfTheKeyIsUnchangedSinceTheRevision(theRevision),
		))
		err = alice.AttemptsTo(
			put.TheValue("3").ForTheKey("app/config/counter").IfTheKeyIsUnchangedSinceTheRevision(theRevision),
		)

		require.ErrorIs(t, err, etcd.ErrKeyModified)
		require.EqualError(
			t,
			err,
			fmt.Sprintf(`the key "app/config/counter" was modified since the revision %d`, theRevision),
		)
		assert.Equal(t, "2", etcdfixture.ValueOf(t, client, "app/config/counter"), "the store is left untouched")
	})

	t.Run("treats the revision zero as the key being absent", func(t *testing.T) {
		require.NoError(t, alice.AttemptsTo(remove.TheKey("app/config/flag")))

		require.NoError(t, alice.AttemptsTo(
			put.TheValue("initial").ForTheKey("app/config/flag").IfTheKeyIsUnchangedSinceTheRevision(0),
		))
		err := alice.AttemptsTo(
			put.TheValue("other").ForTheKey("app/config/flag").IfTheKeyIsUnchangedSinceTheRevision(0),
		)

		require.ErrorIs(t, err, etcd.ErrKeyModified)
	})

	t.Run("attaches the key to a lease when it expires", func(t *testing.T) {
		require.NoError(
			t,
			alice.AttemptsTo(put.TheValue("held").ForTheKey("locks/batch-job").ExpiringAfter(1).Second()),
		)

		theKey, err := alice.AsksFor(key.Named("locks/batch-job"))
		require.NoError(t, err)
		assert.NotZero(t, theKey.Lease())
		assert.Eventually(t, func() bool {
			expired, askErr := alice.AsksFor(key.Named("locks/batch-job"))

			return askErr == nil && !expired.Exists()
		}, 15*time.Second, 200*time.Millisecond, "the key should expire with its lease")
	})

	t.Run("keeps the lease alive until the actor exits", func(t *testing.T) {
		carol := screenplay.ActorNamed("Carol").WhoCan(use.ETCD().WithTheEndpoints(server.Endpoint()))

		require.NoError(t, carol.AttemptsTo(
			put.TheValue("held").ForTheKey("locks/nightly-job").ExpiringAfter(1).Second().AndKeepItAlive(),
		))

		time.Sleep(4 * time.Second)
		assert.Equal(t, "held", etcdfixture.ValueOf(t, client, "locks/nightly-job"), "the lease is renewed")
		require.NoError(t, carol.Exit())
		assert.Eventually(t, func() bool {
			response, err := client.Get(t.Context(), "locks/nightly-job")

			return err == nil && len(response.Kvs) == 0
		}, 15*time.Second, 200*time.Millisecond, "the key should expire once the actor exits")
	})

	t.Run("keeps nothing alive when the key never expires", func(t *testing.T) {
		require.NoError(t, alice.AttemptsTo(put.TheValue("ready").ForTheKey("app/status").AndKeepItAlive()))

		theKey, err := alice.AsksFor(key.Named("app/status"))
		require.NoError(t, err)
		assert.Zero(t, theKey.Lease(), "there is no lease to keep alive")
	})

	t.Run("hides a secret value from the narration only", func(t *testing.T) {
		require.NoError(t, alice.AttemptsTo(put.TheValue("s3cret").ForTheKey("app/config/password").Secretly()))

		assert.Equal(t, "s3cret", etcdfixture.ValueOf(t, client, "app/config/password"))
		assert.Equal(
			t,
			`put the value <secret> for the key "app/config/password"`,
			put.TheValue("s3cret").ForTheKey("app/config/password").Secretly().String(),
		)
	})

	t.Run("reports a missing key before reaching the store", func(t *testing.T) {
		carol := screenplay.ActorNamed("Carol")

		err := carol.AttemptsTo(put.TheValue("v"))
		require.ErrorIs(t, err, etcd.ErrKeyNotProvided)
		require.NotErrorIs(t, err, screenplay.ErrActorMissingAbility)
		require.ErrorIs(t, carol.AttemptsTo(put.TheValue("v").ForTheKey("")), etcd.ErrKeyNotProvided)
	})

	t.Run("fails when the actor cannot use etcd", func(t *testing.T) {
		err := screenplay.ActorNamed("Carol").AttemptsTo(put.TheValue("v").ForTheKey("k"))

		require.ErrorIs(t, err, screenplay.ErrActorMissingAbility)
	})

	t.Run("implements the stringer interface", func(t *testing.T) {
		assert.Equal(t, `put the value "v"`, put.TheValue("v").String())
		assert.Equal(t, `put the value "v" for the key "k"`, put.TheValue("v").ForTheKey("k").String())
		assert.Equal(t,
			`put the value <secret> for the key "k"`,
			put.TheValue("v").ForTheKey("k").WhichShouldBeKeptSecret().String(),
		)
		assert.Equal(t,
			`put the value "v" for the key "k" if the key is absent`,
			put.TheValue("v").ForTheKey("k").IfTheKeyIsAbsent().String(),
		)
		assert.Equal(t,
			`put the value "v" for the key "k" if the key is unchanged since the revision 42`,
			put.TheValue("v").ForTheKey("k").IfTheKeyIsUnchangedSinceTheRevision(42).String(),
		)
		assert.Equal(t,
			`put the value "v" for the key "k" expiring after 2 minutes`,
			put.TheValue("v").ForTheKey("k").ExpiringAfter(2).Minutes().String(),
		)
		assert.Equal(t,
			`put the value "v" for the key "k" expiring after 10 seconds and keep it alive if the key is absent`,
			put.TheValue("v").ForTheKey("k").ExpiringAfter(10).Seconds().AndKeepItAlive().IfTheKeyIsAbsent().String(),
		)
		assert.Equal(t,
			`put the value "v" for the key "k" expiring after 10 seconds and keep it alive if the key is absent`,
			put.TheValue("v").ForTheKey("k").IfTheKeyIsAbsent().AndKeepItAlive().ExpiringAfter(10).Seconds().String(),
			"the narration keeps the same order whatever the order of the settings",
		)
	})

	t.Run("implements the performable interface", func(t *testing.T) {
		assert.Implements(t, (*screenplay.Performable)(nil), put.TheValue("v").ForTheKey("k"))
	})
}
