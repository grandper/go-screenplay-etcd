package ability_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/grandper/go-screenplay/v2/screenplay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grandper/go-screenplay-etcd/etcd"
	"github.com/grandper/go-screenplay-etcd/fixture"
	"github.com/grandper/go-screenplay-etcd/internal/ability"
)

func TestTLS(t *testing.T) {
	secure := fixture.RunSecureETCDServer(t)
	newOperations := func(t *testing.T, theAbility *ability.UseETCDAbility) *ability.Operations {
		t.Helper()
		t.Cleanup(func() { require.NoError(t, theAbility.Forget()) })

		return ability.TheAbilityOf(screenplay.ActorNamed("Alice").WhoCan(theAbility))
	}

	t.Run("secures the connection with the certificate authority and the client certificate", func(t *testing.T) {
		alice := newOperations(t, ability.UseETCD().
			WithTheEndpoints(secure.Endpoint()).
			WithTheCertificateAuthority(secure.CertificateAuthority()).
			WithTheClientCertificate(secure.ClientCertificate()).AndTheKey(secure.ClientKey()).
			WithTimingOutAfter(5).Seconds())

		require.NoError(t, alice.ToPut(t.Context(), "app/status", "ready", 0))

		key, err := alice.ToRetrieveTheKey(t.Context(), "app/status")
		require.NoError(t, err)
		assert.Equal(t, "ready", key.Value())
	})

	t.Run("accepts the certificates in any order", func(t *testing.T) {
		alice := newOperations(t, ability.UseETCD().
			WithTheClientCertificate(secure.ClientCertificate()).AndTheKey(secure.ClientKey()).
			WithTheCertificateAuthority(secure.CertificateAuthority()).
			WithTheEndpoints(secure.Endpoint()).
			WithTimingOutAfter(5).Seconds())

		require.NoError(t, alice.ToPut(t.Context(), "app/status", "ready", 0))
	})

	t.Run("fails without the client certificate the server requires", func(t *testing.T) {
		alice := newOperations(t, ability.UseETCD().
			WithTheEndpoints(secure.Endpoint()).
			WithTheCertificateAuthority(secure.CertificateAuthority()).
			WithTimingOutAfter(2).Seconds())

		require.Error(t, alice.ToPut(t.Context(), "app/status", "ready", 0))
	})

	t.Run("fails when the authority did not sign the certificate of the server", func(t *testing.T) {
		other := fixture.RunSecureETCDServer(t)
		alice := newOperations(t, ability.UseETCD().
			WithTheEndpoints(secure.Endpoint()).
			WithTheCertificateAuthority(other.CertificateAuthority()).
			WithTheClientCertificate(other.ClientCertificate()).AndTheKey(other.ClientKey()).
			WithTimingOutAfter(2).Seconds())

		require.Error(t, alice.ToPut(t.Context(), "app/status", "ready", 0))
	})

	t.Run("reports a certificate authority that cannot be read", func(t *testing.T) {
		alice := newOperations(t, ability.UseETCD().
			WithTheEndpoints(secure.Endpoint()).
			WithTheCertificateAuthority("does/not/exist.crt"))

		err := alice.ToPut(t.Context(), "app/status", "ready", 0)

		require.ErrorIs(t, err, etcd.ErrFailedToConnect)
		require.ErrorIs(t, err, etcd.ErrFailedToLoadTheCertificates)
		require.ErrorIs(t, err, os.ErrNotExist)
	})

	t.Run("reports a certificate authority that holds no certificate", func(t *testing.T) {
		notACertificate := filepath.Join(t.TempDir(), "ca.crt")
		require.NoError(t, os.WriteFile(notACertificate, []byte("not a certificate"), 0o600))
		alice := newOperations(t, ability.UseETCD().
			WithTheEndpoints(secure.Endpoint()).
			WithTheCertificateAuthority(notACertificate))

		err := alice.ToPut(t.Context(), "app/status", "ready", 0)

		require.ErrorIs(t, err, etcd.ErrFailedToLoadTheCertificates)
		require.EqualError(t, err, fmt.Sprintf(
			"failed to connect to the etcd server: failed to load the certificates: %q holds no certificate",
			notACertificate,
		))
	})

	t.Run("reports a client certificate that cannot be read", func(t *testing.T) {
		alice := newOperations(t, ability.UseETCD().
			WithTheEndpoints(secure.Endpoint()).
			WithTheCertificateAuthority(secure.CertificateAuthority()).
			WithTheClientCertificate("does/not/exist.crt").AndTheKey(secure.ClientKey()))

		err := alice.ToPut(t.Context(), "app/status", "ready", 0)

		require.ErrorIs(t, err, etcd.ErrFailedToLoadTheCertificates)
		require.ErrorIs(t, err, os.ErrNotExist)
	})

	t.Run("reports a client certificate whose key cannot be read", func(t *testing.T) {
		alice := newOperations(t, ability.UseETCD().
			WithTheEndpoints(secure.Endpoint()).
			WithTheCertificateAuthority(secure.CertificateAuthority()).
			WithTheClientCertificate(secure.ClientCertificate()).AndTheKey("does/not/exist.key"))

		err := alice.ToPut(t.Context(), "app/status", "ready", 0)

		require.ErrorIs(t, err, etcd.ErrFailedToLoadTheCertificates)
		require.ErrorIs(t, err, os.ErrNotExist)
	})

	t.Run("reports a client certificate that does not match its key", func(t *testing.T) {
		alice := newOperations(t, ability.UseETCD().
			WithTheEndpoints(secure.Endpoint()).
			WithTheCertificateAuthority(secure.CertificateAuthority()).
			WithTheClientCertificate(secure.CertificateAuthority()).AndTheKey(secure.ClientKey()))

		err := alice.ToPut(t.Context(), "app/status", "ready", 0)

		require.ErrorIs(t, err, etcd.ErrFailedToLoadTheCertificates)
	})

	t.Run("reports the certificates when the ability is first used, not when it is given", func(t *testing.T) {
		theAbility := ability.UseETCD().WithTheEndpoints(secure.Endpoint()).WithTheCertificateAuthority("missing.crt")
		alice := newOperations(t, theAbility)

		require.NotNil(t, theAbility, "giving the certificates never fails")
		require.ErrorIs(t, alice.ToPut(t.Context(), "app/status", "ready", 0), etcd.ErrFailedToLoadTheCertificates)
	})

	t.Run("does nothing on a nil receiver", func(t *testing.T) {
		var theAbility *ability.UseETCDAbility

		assert.Nil(t, theAbility.WithTheCertificateAuthority("ca.crt"))
		assert.Nil(t, theAbility.WithTheClientCertificate("client.crt").AndTheKey("client.key"))
	})
}
