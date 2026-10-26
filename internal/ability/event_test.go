package ability_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/grandper/go-screenplay-etcd/internal/ability"
)

func TestEvent(t *testing.T) {
	t.Parallel()

	t.Run("holds the parts of a put that created the key", func(t *testing.T) {
		t.Parallel()

		event := ability.NewEvent(&clientv3.Event{
			Type: mvccpb.PUT,
			Kv: &mvccpb.KeyValue{
				Key:            []byte("app/status"),
				Value:          []byte("ready"),
				Version:        1,
				CreateRevision: 7,
				ModRevision:    7,
			},
		}, false)

		assert.True(t, event.IsPut())
		assert.False(t, event.IsDelete())
		assert.Equal(t, "app/status", event.Key())
		assert.Equal(t, "ready", event.Value())
		assert.Empty(t, event.PreviousValue())
		assert.Equal(t, int64(7), event.Revision())
		assert.Equal(t, int64(1), event.Version())
		assert.False(t, event.CausedByLeaseExpiry())
	})

	t.Run("holds the previous value of a put that overwrote the key", func(t *testing.T) {
		t.Parallel()

		event := ability.NewEvent(&clientv3.Event{
			Type: mvccpb.PUT,
			Kv: &mvccpb.KeyValue{
				Key:            []byte("app/status"),
				Value:          []byte("stopped"),
				Version:        2,
				CreateRevision: 7,
				ModRevision:    8,
			},
			PrevKv: &mvccpb.KeyValue{
				Key:            []byte("app/status"),
				Value:          []byte("ready"),
				Version:        1,
				CreateRevision: 7,
				ModRevision:    7,
			},
		}, false)

		assert.True(t, event.IsPut())
		assert.Equal(t, "stopped", event.Value())
		assert.Equal(t, "ready", event.PreviousValue())
		assert.Equal(t, int64(8), event.Revision())
		assert.Equal(t, int64(2), event.Version())
	})

	t.Run("holds the parts of a delete, whose value and version are gone", func(t *testing.T) {
		t.Parallel()

		event := ability.NewEvent(&clientv3.Event{
			Type: mvccpb.DELETE,
			Kv:   &mvccpb.KeyValue{Key: []byte("app/status"), ModRevision: 9},
			PrevKv: &mvccpb.KeyValue{
				Key:            []byte("app/status"),
				Value:          []byte("stopped"),
				Version:        2,
				CreateRevision: 7,
				ModRevision:    8,
			},
		}, false)

		assert.True(t, event.IsDelete())
		assert.False(t, event.IsPut())
		assert.Equal(t, "app/status", event.Key())
		assert.Empty(t, event.Value())
		assert.Equal(t, "stopped", event.PreviousValue())
		assert.Equal(t, int64(9), event.Revision())
		assert.Zero(t, event.Version())
		assert.False(t, event.CausedByLeaseExpiry())
	})

	t.Run("tells when a delete was caused by the expiry of the lease", func(t *testing.T) {
		t.Parallel()

		event := ability.NewEvent(&clientv3.Event{
			Type:   mvccpb.DELETE,
			Kv:     &mvccpb.KeyValue{Key: []byte("locks/batch-job"), ModRevision: 9},
			PrevKv: &mvccpb.KeyValue{Key: []byte("locks/batch-job"), Value: []byte("held"), Lease: 42},
		}, true)

		assert.True(t, event.IsDelete())
		assert.Equal(t, "held", event.PreviousValue())
		assert.True(t, event.CausedByLeaseExpiry())
	})

	t.Run("holds no previous value when the store sent none", func(t *testing.T) {
		t.Parallel()

		event := ability.NewEvent(&clientv3.Event{
			Type: mvccpb.DELETE,
			Kv:   &mvccpb.KeyValue{Key: []byte("app/status"), ModRevision: 9},
		}, false)

		assert.Empty(t, event.PreviousValue())
	})
}
