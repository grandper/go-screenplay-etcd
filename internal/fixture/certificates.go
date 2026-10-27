package fixture

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// certificateValidity is how long the certificates of a secure server are
// valid: long enough for any test run.
const certificateValidity = 24 * time.Hour

// certificateFileMode is the permission of the certificate and key files:
// readable by their owner only.
const certificateFileMode = 0o600

// certificates holds the files of a certificate authority, of the server
// certificate it signed, and of a client certificate it signed.
type certificates struct {
	caFile         string
	serverCertFile string
	serverKeyFile  string
	clientCertFile string
	clientKeyFile  string
}

// generateCertificates creates a certificate authority, a certificate for a
// server listening on the loopback interface and a certificate for a client,
// both signed by the authority, as PEM files in the directory.
func generateCertificates(t *testing.T, dir string) certificates {
	t.Helper()

	authority := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "go-screenplay-etcd test authority"},
		NotBefore:             time.Now().Add(-time.Minute),
		NotAfter:              time.Now().Add(certificateValidity),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}
	authorityKey := newKey(t)
	authorityPEM := signedCertificate(t, authority, authority, &authorityKey.PublicKey, authorityKey)
	authorityCertificate, err := x509.ParseCertificate(authorityPEM)
	require.NoError(t, err)

	server := &x509.Certificate{
		SerialNumber: big.NewInt(2), //nolint:mnd // the serial numbers only need to differ
		Subject:      pkix.Name{CommonName: "localhost"},
		NotBefore:    time.Now().Add(-time.Minute),
		NotAfter:     time.Now().Add(certificateValidity),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1)}, //nolint:mnd // the loopback address
		DNSNames:     []string{"localhost"},
	}
	serverKey := newKey(t)
	serverPEM := signedCertificate(t, server, authorityCertificate, &serverKey.PublicKey, authorityKey)

	client := &x509.Certificate{
		SerialNumber: big.NewInt(3), //nolint:mnd // the serial numbers only need to differ
		Subject:      pkix.Name{CommonName: "go-screenplay-etcd test client"},
		NotBefore:    time.Now().Add(-time.Minute),
		NotAfter:     time.Now().Add(certificateValidity),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	clientKey := newKey(t)
	clientPEM := signedCertificate(t, client, authorityCertificate, &clientKey.PublicKey, authorityKey)

	return certificates{
		caFile:         writePEM(t, filepath.Join(dir, "ca.crt"), "CERTIFICATE", authorityPEM),
		serverCertFile: writePEM(t, filepath.Join(dir, "server.crt"), "CERTIFICATE", serverPEM),
		serverKeyFile:  writeKey(t, filepath.Join(dir, "server.key"), serverKey),
		clientCertFile: writePEM(t, filepath.Join(dir, "client.crt"), "CERTIFICATE", clientPEM),
		clientKeyFile:  writeKey(t, filepath.Join(dir, "client.key"), clientKey),
	}
}

// newKey generates a private key.
func newKey(t *testing.T) *ecdsa.PrivateKey {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	return key
}

// signedCertificate creates the certificate from its template, signed by the
// parent with the key of the parent, in DER form.
func signedCertificate(
	t *testing.T,
	template, parent *x509.Certificate,
	publicKey *ecdsa.PublicKey,
	parentKey *ecdsa.PrivateKey,
) []byte {
	t.Helper()

	der, err := x509.CreateCertificate(rand.Reader, template, parent, publicKey, parentKey)
	require.NoError(t, err)

	return der
}

// writePEM writes the DER bytes to the file as a PEM block of the given type,
// and returns the path of the file.
func writePEM(t *testing.T, path, blockType string, der []byte) string {
	t.Helper()

	content := pem.EncodeToMemory(&pem.Block{Type: blockType, Bytes: der})
	require.NoError(t, os.WriteFile(path, content, certificateFileMode))

	return path
}

// writeKey writes the private key to the file as a PEM block, and returns the
// path of the file.
func writeKey(t *testing.T, path string, key *ecdsa.PrivateKey) string {
	t.Helper()

	der, err := x509.MarshalECPrivateKey(key)
	require.NoError(t, err)

	return writePEM(t, path, "EC PRIVATE KEY", der)
}
