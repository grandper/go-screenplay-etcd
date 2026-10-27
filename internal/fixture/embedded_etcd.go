// Package fixture spins up an in-process etcd server, so that scenarios run
// without an external cluster:
//
//	server := fixture.RunETCDServer(t)
//	anActor := screenplay.ActorNamed("Alice").WhoCan(use.ETCD().WithTheEndpoints(server.Endpoint()))
package fixture

import (
	"crypto/tls"
	"crypto/x509"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.etcd.io/etcd/client/pkg/v3/transport"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/server/v3/embed"
)

// startupTimeout bounds the time the server is given to elect itself leader
// before the test fails.
const startupTimeout = 30 * time.Second

// dialTimeout bounds the time a client of the fixture is given to reach the
// server.
const dialTimeout = 5 * time.Second

// ETCDServer is an etcd server running in the test process. It listens on a
// free port of the loopback interface, keeps its data in the temporary
// directory of the test, and is stopped when the test ends.
type ETCDServer struct {
	etcd         *embed.Etcd
	logFile      string
	certificates *certificates
}

// RunETCDServer starts an etcd server for the test, and registers its shutdown
// with the test's cleanup. Every call starts its own server, on its own port,
// so that tests never share keys.
func RunETCDServer(t *testing.T) *ETCDServer {
	t.Helper()

	return runETCDServer(t, nil)
}

// RunSecureETCDServer starts an etcd server that secures its connections with
// TLS and requires its clients to present a certificate, as RunETCDServer
// does. It generates a certificate authority, a server certificate and a
// client certificate for the test; the actor connects with the files
// CertificateAuthority, ClientCertificate and ClientKey return:
//
//	server := fixture.RunSecureETCDServer(t)
//	anActor := screenplay.ActorNamed("Alice").WhoCan(
//		use.ETCD().
//			WithTheEndpoints(server.Endpoint()).
//			WithTheCertificateAuthority(server.CertificateAuthority()).
//			WithTheClientCertificate(server.ClientCertificate()).AndTheKey(server.ClientKey()),
//	)
func RunSecureETCDServer(t *testing.T) *ETCDServer {
	t.Helper()

	generated := generateCertificates(t, t.TempDir())

	return runETCDServer(t, &generated)
}

// runETCDServer starts the server, secured with the certificates when they are
// given.
func runETCDServer(t *testing.T, certificates *certificates) *ETCDServer {
	t.Helper()

	dir := t.TempDir()
	config := embed.NewConfig()
	config.Dir = filepath.Join(dir, "data")
	config.LogLevel = "error"
	config.LogOutputs = []string{filepath.Join(dir, "etcd.log")}

	listenOnAFreePort(t, config, certificates != nil)

	if certificates != nil {
		config.ClientTLSInfo = transport.TLSInfo{
			CertFile:       certificates.serverCertFile,
			KeyFile:        certificates.serverKeyFile,
			TrustedCAFile:  certificates.caFile,
			ClientCertAuth: true,
		}
	}

	etcd, err := embed.StartEtcd(config)
	require.NoError(t, err)
	t.Cleanup(func() {
		etcd.Close()
		<-etcd.Server.StopNotify()
	})

	select {
	case <-etcd.Server.ReadyNotify():
	case <-time.After(startupTimeout):
		t.Fatal("the embedded etcd server did not become ready in time")
	}

	return &ETCDServer{
		etcd:         etcd,
		logFile:      filepath.Join(dir, "etcd.log"),
		certificates: certificates,
	}
}

// listenOnAFreePort makes the server listen for its clients and its peers on a
// port chosen by the system, so that several servers can run at once.
func listenOnAFreePort(t *testing.T, config *embed.Config, secure bool) {
	t.Helper()

	scheme := "http"
	if secure {
		scheme = "https"
	}

	clientURL, err := url.Parse(scheme + "://127.0.0.1:0")
	require.NoError(t, err)
	peerURL, err := url.Parse("http://127.0.0.1:0")
	require.NoError(t, err)

	config.ListenClientUrls = []url.URL{*clientURL}
	config.AdvertiseClientUrls = []url.URL{*clientURL}
	config.ListenPeerUrls = []url.URL{*peerURL}
	config.AdvertisePeerUrls = []url.URL{*peerURL}
	config.InitialCluster = config.InitialClusterFromName(config.Name)
}

// Endpoint returns the address the server listens on for its clients, in the
// form host:port.
func (s *ETCDServer) Endpoint() string {
	return s.etcd.Clients[0].Addr().String()
}

// NewClient returns a client connected to the server, closed when the test
// ends. It is meant for checking the state of the store from outside the
// actor. The client of a secure server presents the client certificate.
func (s *ETCDServer) NewClient(t *testing.T) *clientv3.Client {
	t.Helper()

	config := clientv3.Config{
		Endpoints:   []string{s.Endpoint()},
		DialTimeout: dialTimeout,
	}
	if s.certificates != nil {
		config.TLS = s.clientTLSConfig(t)
	}

	client, err := clientv3.New(config)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, client.Close())
	})

	return client
}

// clientTLSConfig builds the TLS configuration of a client of the secure
// server: it trusts the authority and presents the client certificate.
func (s *ETCDServer) clientTLSConfig(t *testing.T) *tls.Config {
	t.Helper()

	authority, err := os.ReadFile(s.certificates.caFile)
	require.NoError(t, err)
	pool := x509.NewCertPool()
	require.True(t, pool.AppendCertsFromPEM(authority))

	certificate, err := tls.LoadX509KeyPair(s.certificates.clientCertFile, s.certificates.clientKeyFile)
	require.NoError(t, err)

	return &tls.Config{
		MinVersion:   tls.VersionTLS12,
		RootCAs:      pool,
		Certificates: []tls.Certificate{certificate},
	}
}

// LogFile returns the path of the file the server writes its logs to, so that
// they never mix with the output of the test.
func (s *ETCDServer) LogFile() string {
	return s.logFile
}

// CertificateAuthority returns the path of the PEM file of the authority that
// signed the certificates of a secure server, empty for a plain one.
func (s *ETCDServer) CertificateAuthority() string {
	if s.certificates == nil {
		return ""
	}

	return s.certificates.caFile
}

// ClientCertificate returns the path of the PEM file of the certificate a
// client of a secure server presents, empty for a plain one.
func (s *ETCDServer) ClientCertificate() string {
	if s.certificates == nil {
		return ""
	}

	return s.certificates.clientCertFile
}

// ClientKey returns the path of the PEM file of the private key of the client
// certificate of a secure server, empty for a plain one.
func (s *ETCDServer) ClientKey() string {
	if s.certificates == nil {
		return ""
	}

	return s.certificates.clientKeyFile
}
