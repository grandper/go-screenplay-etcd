package put_test

import (
	"testing"

	"github.com/grandper/go-screenplay/v2/screenplay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/grandper/go-screenplay-etcd/ability/use"
	"github.com/grandper/go-screenplay-etcd/etcd"
	"github.com/grandper/go-screenplay-etcd/internal/action/put"
	"github.com/grandper/go-screenplay-etcd/internal/fixture"
)

func TestTheValues(t *testing.T) {
	server := fixture.RunETCDServer(t)
	client := server.NewClient(t)
	alice := screenplay.ActorNamed("Alice").WhoCan(use.ETCD().WithTheEndpoints(server.Endpoint()))

	t.Run("puts every value for its key at the same revision", func(t *testing.T) {
		require.NoError(t, alice.AttemptsTo(put.TheValues(map[string]string{
			"app/config/database": "postgresql://localhost:5432",
			"app/config/cache":    "redis://localhost:6379",
		})))

		assert.Equal(t, "postgresql://localhost:5432", fixture.ValueOf(t, client, "app/config/database"))
		assert.Equal(t, "redis://localhost:6379", fixture.ValueOf(t, client, "app/config/cache"))
		response, err := client.Get(t.Context(), "app/config/", clientv3.WithPrefix())
		require.NoError(t, err)
		require.Len(t, response.Kvs, 2)
		assert.Equal(t, response.Kvs[0].GetModRevision(), response.Kvs[1].GetModRevision(), "one transaction")
	})

	t.Run("puts one value", func(t *testing.T) {
		require.NoError(t, alice.AttemptsTo(put.TheValues(map[string]string{"app/status": "ready"})))

		assert.Equal(t, "ready", fixture.ValueOf(t, client, "app/status"))
	})

	t.Run("reports missing values before reaching the store", func(t *testing.T) {
		carol := screenplay.ActorNamed("Carol")

		require.ErrorIs(t, carol.AttemptsTo(put.TheValues(nil)), etcd.ErrKeyNotProvided)
		require.ErrorIs(t, carol.AttemptsTo(put.TheValues(map[string]string{})), etcd.ErrKeyNotProvided)
	})

	t.Run("reports an empty key before reaching the store", func(t *testing.T) {
		err := screenplay.ActorNamed("Carol").AttemptsTo(put.TheValues(map[string]string{"k": "v", "": "v"}))

		require.ErrorIs(t, err, etcd.ErrKeyNotProvided)
		require.NotErrorIs(t, err, screenplay.ErrActorMissingAbility)
	})

	t.Run("fails when the actor cannot use etcd", func(t *testing.T) {
		err := screenplay.ActorNamed("Carol").AttemptsTo(put.TheValues(map[string]string{"k": "v"}))

		require.ErrorIs(t, err, screenplay.ErrActorMissingAbility)
	})

	t.Run("implements the stringer interface, with the values sorted by key", func(t *testing.T) {
		values := map[string]string{
			"app/config/database": "postgresql://localhost:5432",
			"app/config/cache":    "redis://localhost:6379",
		}

		assert.Equal(t,
			`put the values "app/config/cache" = "redis://localhost:6379", `+
				`"app/config/database" = "postgresql://localhost:5432"`,
			put.TheValues(values).String(),
		)
		assert.Equal(t,
			`put the values <secret> for the keys "app/config/cache", "app/config/database"`,
			put.TheValues(values).Secretly().String(),
		)
		assert.Equal(t,
			`put the values <secret> for the keys "app/config/cache", "app/config/database"`,
			put.TheValues(values).WhichShouldBeKeptSecret().String(),
		)
		assert.Equal(t, "put the values", put.TheValues(nil).String())
		assert.Equal(t, "put the values", put.TheValues(map[string]string{}).Secretly().String())
	})

	t.Run("implements the performable interface", func(t *testing.T) {
		assert.Implements(t, (*screenplay.Performable)(nil), put.TheValues(map[string]string{"k": "v"}))
	})
}
