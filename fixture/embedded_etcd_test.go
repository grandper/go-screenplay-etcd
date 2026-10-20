package fixture_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/grandper/go-screenplay-etcd/fixture"
)

func TestRunETCDServer(t *testing.T) {
	t.Run("starts a server that accepts puts and gets", func(t *testing.T) {
		server := fixture.RunETCDServer(t)
		assert.NotEmpty(t, server.Endpoint())

		client := server.NewClient(t)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := client.Put(ctx, "greeting", "hello")
		require.NoError(t, err)

		response, err := client.Get(ctx, "greeting")
		require.NoError(t, err)
		require.Len(t, response.Kvs, 1)
		assert.Equal(t, "hello", string(response.Kvs[0].GetValue()))
	})

	t.Run("starts servers that do not conflict with one another", func(t *testing.T) {
		first := fixture.RunETCDServer(t)
		second := fixture.RunETCDServer(t)

		assert.NotEqual(t, first.Endpoint(), second.Endpoint())
	})

	t.Run("keeps its logs out of the output of the test", func(t *testing.T) {
		server := fixture.RunETCDServer(t)

		assert.FileExists(t, server.LogFile())
	})
}

func TestRunSecureETCDServer(t *testing.T) {
	t.Run("starts a server that requires a client certificate signed by its authority", func(t *testing.T) {
		server := fixture.RunSecureETCDServer(t)
		assert.FileExists(t, server.CertificateAuthority())
		assert.FileExists(t, server.ClientCertificate())
		assert.FileExists(t, server.ClientKey())

		client := server.NewClient(t)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := client.Put(ctx, "greeting", "hello")
		require.NoError(t, err)
	})

	t.Run("refuses a client without a certificate", func(t *testing.T) {
		server := fixture.RunSecureETCDServer(t)

		client, err := clientv3.New(clientv3.Config{Endpoints: []string{server.Endpoint()}})
		require.NoError(t, err)
		t.Cleanup(func() { require.NoError(t, client.Close()) })
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		_, err = client.Put(ctx, "greeting", "hello")
		require.Error(t, err)
	})
}
