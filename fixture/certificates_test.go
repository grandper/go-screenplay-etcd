package fixture_test

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"net"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grandper/go-screenplay-etcd/fixture"
)

// parseCertificate parses the certificate held by the PEM file.
func parseCertificate(t *testing.T, path string) *x509.Certificate {
	t.Helper()

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	block, rest := pem.Decode(content)
	require.NotNil(t, block, "%q holds no PEM block", path)
	require.Empty(t, rest, "%q holds more than one PEM block", path)
	require.Equal(t, "CERTIFICATE", block.Type)

	certificate, err := x509.ParseCertificate(block.Bytes)
	require.NoError(t, err)

	return certificate
}

// clientTLSConfig builds the configuration of a client that trusts the
// authority of the server and presents its client certificate.
func clientTLSConfig(t *testing.T, server *fixture.ETCDServer) *tls.Config {
	t.Helper()

	pool := x509.NewCertPool()
	pool.AddCert(parseCertificate(t, server.CertificateAuthority()))
	certificate, err := tls.LoadX509KeyPair(server.ClientCertificate(), server.ClientKey())
	require.NoError(t, err)

	return &tls.Config{MinVersion: tls.VersionTLS12, RootCAs: pool, Certificates: []tls.Certificate{certificate}}
}

func TestCertificates(t *testing.T) {
	server := fixture.RunSecureETCDServer(t)

	t.Run("include a certificate authority of its own", func(t *testing.T) {
		authority := parseCertificate(t, server.CertificateAuthority())

		assert.True(t, authority.IsCA)
		assert.Equal(t, "go-screenplay-etcd test authority", authority.Subject.CommonName)
		require.NoError(t, authority.CheckSignatureFrom(authority), "the authority signs itself")
	})

	t.Run("include a client certificate signed by the authority, with its key", func(t *testing.T) {
		authority := parseCertificate(t, server.CertificateAuthority())
		client := parseCertificate(t, server.ClientCertificate())

		assert.Equal(t, "go-screenplay-etcd test client", client.Subject.CommonName)
		require.NoError(t, client.CheckSignatureFrom(authority))
		assert.Contains(t, client.ExtKeyUsage, x509.ExtKeyUsageClientAuth)
		_, err := tls.LoadX509KeyPair(server.ClientCertificate(), server.ClientKey())
		require.NoError(t, err, "the key matches the certificate")
	})

	t.Run("include a server certificate signed by the authority, valid for the loopback interface", func(t *testing.T) {
		connection, err := tls.Dial("tcp", server.Endpoint(), clientTLSConfig(t, server))
		require.NoError(t, err, "the client trusts the server at its address")
		t.Cleanup(func() { require.NoError(t, connection.Close()) })

		certificates := connection.ConnectionState().PeerCertificates
		require.NotEmpty(t, certificates)
		assert.Equal(t, "localhost", certificates[0].Subject.CommonName)
		assert.Contains(t, certificates[0].DNSNames, "localhost")
		assert.True(t, certificates[0].IPAddresses[0].Equal(net.IPv4(127, 0, 0, 1)))
		require.NoError(t, certificates[0].CheckSignatureFrom(parseCertificate(t, server.CertificateAuthority())))
	})

	t.Run("are readable by their owner only", func(t *testing.T) {
		for _, path := range []string{server.CertificateAuthority(), server.ClientCertificate(), server.ClientKey()} {
			info, err := os.Stat(path)
			require.NoError(t, err)
			assert.Equal(t, os.FileMode(0o600), info.Mode().Perm(), "%q", path)
		}
	})

	t.Run("are generated for every server, so that one never trusts another", func(t *testing.T) {
		other := fixture.RunSecureETCDServer(t)

		assert.NotEqual(t, server.CertificateAuthority(), other.CertificateAuthority(), "the files differ")
		require.False(
			t,
			parseCertificate(t, server.CertificateAuthority()).Equal(parseCertificate(t, other.CertificateAuthority())),
			"the authorities differ",
		)
		_, err := tls.Dial("tcp", other.Endpoint(), clientTLSConfig(t, server))
		require.Error(t, err, "the client of one server is refused by the other")
	})
}
