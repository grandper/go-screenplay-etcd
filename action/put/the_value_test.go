package put_test

import (
	"fmt"
	"testing"

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
	"github.com/grandper/go-screenplay-etcd/question"
)

func TestTheValue(t *testing.T) {
	server := fixture.RunETCDServer(t)
	client := server.NewClient(t)
	alice := screenplay.ActorNamed("Alice").WhoCan(ability.UseETCD().WithTheEndpoints(server.Endpoint()))

	t.Run("puts a value", func(t *testing.T) {
		t.Run("for a key", func(t *testing.T) {
			require.NoError(
				t,
				alice.AttemptsTo(put.TheValue("postgresql://localhost:5432").ForTheKey("app/config/database")),
			)

			assert.Equal(t, "postgresql://localhost:5432", etcdfixture.ValueOf(t, client, "app/config/database"))
		})

		t.Run("overwriting the value the key already holds", func(t *testing.T) {
			require.NoError(t, alice.AttemptsTo(
				put.TheValue("1").ForTheKey("app/config/retries"),
				put.TheValue("2").ForTheKey("app/config/retries"),
			))

			assert.Equal(t, "2", etcdfixture.ValueOf(t, client, "app/config/retries"))
		})

		t.Run("that is empty", func(t *testing.T) {
			require.NoError(t, alice.AttemptsTo(put.TheValue("").ForTheKey("app/config/empty")))

			assert.Empty(t, etcdfixture.ValueOf(t, client, "app/config/empty"))
		})

		t.Run("only if the key is absent", func(t *testing.T) {
			require.NoError(t, alice.AttemptsTo(remove.TheKey("app/config/flag")))

			require.NoError(
				t,
				alice.AttemptsTo(put.TheValue("initial").ForTheKey("app/config/flag").IfTheKeyIsAbsent()),
			)
			assert.Equal(t, "initial", etcdfixture.ValueOf(t, client, "app/config/flag"))

			err := alice.AttemptsTo(put.TheValue("other").ForTheKey("app/config/flag").IfTheKeyIsAbsent())
			require.ErrorIs(t, err, action.ErrKeyAlreadyExists)
			require.EqualError(t, err, `the key "app/config/flag" already exists`)
			assert.Equal(t, "initial", etcdfixture.ValueOf(t, client, "app/config/flag"))
		})

		t.Run("only if the key is unchanged since a revision", func(t *testing.T) {
			require.NoError(t, alice.AttemptsTo(put.TheValue("1").ForTheKey("app/config/counter")))
			revision, err := alice.AsksFor(question.RevisionOfTheKey("app/config/counter"))
			require.NoError(t, err)

			require.NoError(t, alice.AttemptsTo(
				put.TheValue("2").ForTheKey("app/config/counter").IfTheKeyIsUnchangedSinceTheRevision(revision),
			))
			assert.Equal(t, "2", etcdfixture.ValueOf(t, client, "app/config/counter"))

			err = alice.AttemptsTo(
				put.TheValue("3").ForTheKey("app/config/counter").IfTheKeyIsUnchangedSinceTheRevision(revision),
			)
			require.ErrorIs(t, err, action.ErrKeyModified)
			require.EqualError(t, err,
				fmt.Sprintf(`the key "app/config/counter" was modified since the revision %d`, revision),
			)
			assert.Equal(t, "2", etcdfixture.ValueOf(t, client, "app/config/counter"))
		})

		t.Run("only if the key is unchanged since the revision zero, which means absent", func(t *testing.T) {
			require.NoError(t, alice.AttemptsTo(remove.TheKey("app/config/flag")))

			require.NoError(t, alice.AttemptsTo(
				put.TheValue("initial").ForTheKey("app/config/flag").IfTheKeyIsUnchangedSinceTheRevision(0),
			))

			require.ErrorIs(
				t,
				alice.AttemptsTo(
					put.TheValue("other").ForTheKey("app/config/flag").IfTheKeyIsUnchangedSinceTheRevision(0),
				),
				action.ErrKeyModified,
			)
		})

		t.Run("fails when the key is not provided", func(t *testing.T) {
			err := alice.AttemptsTo(put.TheValue("v"))
			require.ErrorIs(t, err, action.ErrKeyNotProvided)
			require.EqualError(t, err, "the key must be provided")

			require.ErrorIs(t, alice.AttemptsTo(put.TheValue("v").ForTheKey("")), action.ErrKeyNotProvided)
		})

		t.Run("fails when the actor cannot use etcd", func(t *testing.T) {
			bob := screenplay.ActorNamed("Bob")

			require.ErrorIs(t, bob.AttemptsTo(put.TheValue("v").ForTheKey("k")), screenplay.ErrActorMissingAbility)
		})
	})

	t.Run("is narrated", func(t *testing.T) {
		recorder := screenplayfixture.NewRecorder()
		adam := etcdfixture.NarratedActor("Adam", recorder, server.Endpoint())

		require.NoError(t, adam.AttemptsTo(put.TheValue("redis://localhost:6379").ForTheKey("app/config/cache")))

		assert.Equal(t,
			`Adam attempts to put the value "redis://localhost:6379" for the key "app/config/cache"`,
			recorder.Messages()[0],
		)
	})

	t.Run("can be kept secret", func(t *testing.T) {
		recorder := screenplayfixture.NewRecorder()
		adam := etcdfixture.NarratedActor("Adam", recorder, server.Endpoint())

		require.NoError(t, adam.AttemptsTo(put.TheValue("s3cret").ForTheKey("app/config/password").Secretly()))

		assert.Equal(
			t,
			`Adam attempts to put the value <secret> for the key "app/config/password"`,
			recorder.Messages()[0],
		)
		assert.Equal(t, "s3cret", etcdfixture.ValueOf(t, client, "app/config/password"))
	})

	t.Run("implements the stringer interface", func(t *testing.T) {
		assert.Equal(t,
			`put the value "postgresql://localhost:5432" for the key "app/config/database"`,
			put.TheValue("postgresql://localhost:5432").ForTheKey("app/config/database").String(),
		)
		assert.Equal(t, `put the value "v"`, put.TheValue("v").String())
		assert.Equal(t, `put the value <secret> for the key "k"`, put.TheValue("v").ForTheKey("k").Secretly().String())
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
	})
}
