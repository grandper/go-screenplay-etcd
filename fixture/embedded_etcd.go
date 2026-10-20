// Package fixture spins up an in-process etcd server, so that scenarios run
// without an external cluster:
//
//	server := fixture.RunETCDServer(t)
//	anActor := screenplay.ActorNamed("Alice").WhoCan(ability.UseETCD().WithTheEndpoints(server.Endpoint()))
package fixture

import (
	"net/url"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
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
	etcd    *embed.Etcd
	logFile string
}

// RunETCDServer starts an etcd server for the test, and registers its shutdown
// with the test's cleanup. Every call starts its own server, on its own port,
// so that tests never share keys.
func RunETCDServer(t *testing.T) *ETCDServer {
	t.Helper()

	dir := t.TempDir()
	config := embed.NewConfig()
	config.Dir = filepath.Join(dir, "data")
	config.LogLevel = "error"
	config.LogOutputs = []string{filepath.Join(dir, "etcd.log")}

	listenOnAFreePort(t, config)

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
		etcd:    etcd,
		logFile: filepath.Join(dir, "etcd.log"),
	}
}

// listenOnAFreePort makes the server listen for its clients and its peers on a
// port chosen by the system, so that several servers can run at once.
func listenOnAFreePort(t *testing.T, config *embed.Config) {
	t.Helper()

	clientURL, err := url.Parse("http://127.0.0.1:0")
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
// actor.
func (s *ETCDServer) NewClient(t *testing.T) *clientv3.Client {
	t.Helper()

	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{s.Endpoint()},
		DialTimeout: dialTimeout,
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, client.Close())
	})

	return client
}

// LogFile returns the path of the file the server writes its logs to, so that
// they never mix with the output of the test.
func (s *ETCDServer) LogFile() string {
	return s.logFile
}
