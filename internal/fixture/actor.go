package fixture

import (
	"testing"

	"github.com/grandper/go-screenplay/v2/screenplay"
	screenplayfixture "github.com/grandper/go-screenplay/v2/screenplay/fixture"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/grandper/go-screenplay-etcd/ability/use"
)

// NarratedActor creates an actor whose narration is recorded by the recorder,
// and who can use the etcd server at the endpoint.
func NarratedActor(name string, recorder *screenplayfixture.Recorder, endpoint string) *screenplay.Actor {
	production := screenplay.NewProduction(screenplay.WithNarrator(screenplay.NewNarrator(recorder)))

	return production.SetTheStage(screenplay.CastOfStandardActors()).
		ActorNamed(name).
		WhoCan(use.ETCD().WithTheEndpoints(endpoint))
}

// ValueOf returns the value the key holds, read with the client, and fails
// the test when the key does not exist.
func ValueOf(t *testing.T, client *clientv3.Client, key string) string {
	t.Helper()

	response, err := client.Get(t.Context(), key)
	require.NoError(t, err)
	require.Len(t, response.Kvs, 1, "the key %q does not exist", key)

	return string(response.Kvs[0].GetValue())
}

// AssertKeyDoesNotExist asserts that the key does not exist in the store the
// client is connected to.
func AssertKeyDoesNotExist(t *testing.T, client *clientv3.Client, key string) {
	t.Helper()

	response, err := client.Get(t.Context(), key)
	require.NoError(t, err)
	assert.Empty(t, response.Kvs, "the key %q exists", key)
}

// RevisionOf returns the current revision of the store the client is connected
// to.
func RevisionOf(t *testing.T, client *clientv3.Client) int64 {
	t.Helper()

	response, err := client.Get(t.Context(), "revision", clientv3.WithCountOnly())
	require.NoError(t, err)

	return response.Header.GetRevision()
}
