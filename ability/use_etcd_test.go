package ability_test

import (
	"context"
	"testing"
	"time"

	"github.com/grandper/go-screenplay/v2/screenplay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grandper/go-screenplay-etcd/ability"
	"github.com/grandper/go-screenplay-etcd/action/put"
	"github.com/grandper/go-screenplay-etcd/fixture"
	"github.com/grandper/go-screenplay-etcd/question"
)

func TestUseETCD(t *testing.T) {
	server := fixture.RunETCDServer(t)

	t.Run("gives an actor the ability to use etcd", func(t *testing.T) {
		alice := screenplay.ActorNamed("Alice").WhoCan(ability.UseETCD().WithTheEndpoints(server.Endpoint()))

		require.NoError(
			t,
			alice.AttemptsTo(put.TheValue("postgresql://localhost:5432").ForTheKey("app/config/database")),
		)

		value, err := alice.AsksFor(question.ValueOfTheKey("app/config/database"))
		require.NoError(t, err)
		assert.Equal(t, "postgresql://localhost:5432", value)
	})

	t.Run("reports a missing endpoint when the actor first uses the ability", func(t *testing.T) {
		alice := screenplay.ActorNamed("Alice").WhoCan(ability.UseETCD())

		err := alice.AttemptsTo(put.TheValue("v").ForTheKey("k"))

		require.ErrorIs(t, err, ability.ErrNoEndpointProvided)
	})

	t.Run("bounds every operation with the timeout", func(t *testing.T) {
		// Nothing listens on port 1, so the operation waits until the timeout.
		alice := screenplay.ActorNamed("Alice").WhoCan(
			ability.UseETCD().WithTheEndpoints("127.0.0.1:1").WithTimingOutAfter(200).Milliseconds(),
		)

		started := time.Now()
		err := alice.AttemptsTo(put.TheValue("v").ForTheKey("k"))

		require.ErrorIs(t, err, context.DeadlineExceeded)
		assert.Less(t, time.Since(started), 5*time.Second)
	})

	t.Run("stops an operation when the context of the actor is done", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()
		alice := screenplay.ActorNamed("Alice").WhoCan(ability.UseETCD().WithTheEndpoints("127.0.0.1:1"))

		err := alice.WithContext(ctx).AttemptsTo(put.TheValue("v").ForTheKey("k"))

		require.ErrorIs(t, err, context.DeadlineExceeded)
	})

	t.Run("prefixes every key with the namespace", func(t *testing.T) {
		alice := screenplay.ActorNamed("Alice").WhoCan(
			ability.UseETCD().WithTheEndpoints(server.Endpoint()).WithTheNamespace("test-run-42/"),
		)

		require.NoError(t, alice.AttemptsTo(put.TheValue("v1").ForTheKey("app/config/database")))

		response, err := server.NewClient(t).Get(t.Context(), "test-run-42/app/config/database")
		require.NoError(t, err)
		require.Len(t, response.Kvs, 1)
		assert.Equal(t, "v1", string(response.Kvs[0].GetValue()))

		value, err := alice.AsksFor(question.ValueOfTheKey("app/config/database"))
		require.NoError(t, err)
		assert.Equal(t, "v1", value)
	})

	t.Run("uses the client it is given", func(t *testing.T) {
		client := server.NewClient(t)
		alice := screenplay.ActorNamed("Alice").WhoCan(ability.UseETCD().WithTheClient(client))

		require.NoError(t, alice.AttemptsTo(put.TheValue("v2").ForTheKey("app/config/cache")))

		response, err := client.Get(t.Context(), "app/config/cache")
		require.NoError(t, err)
		require.Len(t, response.Kvs, 1)
		assert.Equal(t, "v2", string(response.Kvs[0].GetValue()))
	})

	t.Run("closes the connection when the actor exits", func(t *testing.T) {
		alice := screenplay.ActorNamed("Alice").WhoCan(ability.UseETCD().WithTheEndpoints(server.Endpoint()))
		require.NoError(t, alice.AttemptsTo(put.TheValue("v").ForTheKey("k")))

		require.NoError(t, alice.Exit())

		require.ErrorIs(t, alice.AttemptsTo(put.TheValue("v").ForTheKey("k")), screenplay.ErrActorMissingAbility)
	})

	t.Run("exits without an error when the actor never used the ability", func(t *testing.T) {
		alice := screenplay.ActorNamed("Alice").WhoCan(ability.UseETCD().WithTheEndpoints(server.Endpoint()))

		require.NoError(t, alice.Exit())
	})

	t.Run("connects securely with a client certificate and its key", func(t *testing.T) {
		secure := fixture.RunSecureETCDServer(t)
		alice := screenplay.ActorNamed("Alice").WhoCan(
			ability.UseETCD().
				WithTheEndpoints(secure.Endpoint()).
				WithTheCertificateAuthority(secure.CertificateAuthority()).
				WithTheClientCertificate(secure.ClientCertificate()).AndTheKey(secure.ClientKey()).
				WithTimingOutAfter(5).Seconds(),
		)

		require.NoError(t, alice.AttemptsTo(put.TheValue("v").ForTheKey("k")))

		value, err := alice.AsksFor(question.ValueOfTheKey("k"))
		require.NoError(t, err)
		assert.Equal(t, "v", value)
	})

	t.Run("fails to connect securely without the client certificate the server requires", func(t *testing.T) {
		secure := fixture.RunSecureETCDServer(t)
		alice := screenplay.ActorNamed("Alice").WhoCan(
			ability.UseETCD().
				WithTheEndpoints(secure.Endpoint()).
				WithTheCertificateAuthority(secure.CertificateAuthority()).
				WithTimingOutAfter(2).Seconds(),
		)

		require.Error(t, alice.AttemptsTo(put.TheValue("v").ForTheKey("k")))
	})

	t.Run("reports a certificate that cannot be loaded when the actor first uses the ability", func(t *testing.T) {
		alice := screenplay.ActorNamed("Alice").WhoCan(
			ability.UseETCD().
				WithTheEndpoints(server.Endpoint()).
				WithTheClientCertificate("does/not/exist.crt").AndTheKey("does/not/exist.key"),
		)

		err := alice.AttemptsTo(put.TheValue("v").ForTheKey("k"))

		require.ErrorIs(t, err, ability.ErrFailedToLoadTheCertificates)
		require.ErrorIs(t, err, ability.ErrFailedToConnect)
	})

	t.Run("reports a certificate authority that cannot be loaded", func(t *testing.T) {
		alice := screenplay.ActorNamed("Alice").WhoCan(
			ability.UseETCD().WithTheEndpoints(server.Endpoint()).WithTheCertificateAuthority("does/not/exist.crt"),
		)

		err := alice.AttemptsTo(put.TheValue("v").ForTheKey("k"))

		require.ErrorIs(t, err, ability.ErrFailedToLoadTheCertificates)
	})

	t.Run("reports a certificate authority that is not a certificate", func(t *testing.T) {
		alice := screenplay.ActorNamed("Alice").WhoCan(
			ability.UseETCD().WithTheEndpoints(server.Endpoint()).WithTheCertificateAuthority(server.LogFile()),
		)

		err := alice.AttemptsTo(put.TheValue("v").ForTheKey("k"))

		require.ErrorIs(t, err, ability.ErrFailedToLoadTheCertificates)
	})

	t.Run("accepts its settings in any order", func(t *testing.T) {
		alice := screenplay.ActorNamed("Alice").WhoCan(
			ability.UseETCD().
				WithTheNamespace("ordered/").
				WithTimingOutAfter(5).Seconds().
				WithTheEndpoints(server.Endpoint()),
		)

		require.NoError(t, alice.AttemptsTo(put.TheValue("v").ForTheKey("k")))
	})
}
