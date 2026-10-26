package ability_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.etcd.io/etcd/api/v3/mvccpb"

	"github.com/grandper/go-screenplay-etcd/internal/ability"
)

func TestKey(t *testing.T) {
	t.Parallel()

	t.Run("holds the parts of the key the store answered with", func(t *testing.T) {
		t.Parallel()

		key := ability.NewKey(&mvccpb.KeyValue{
			Key:            []byte("app/config/database"),
			Value:          []byte("postgresql://localhost:5432"),
			Version:        3,
			CreateRevision: 10,
			ModRevision:    12,
			Lease:          42,
		})

		assert.True(t, key.Exists())
		assert.Equal(t, "app/config/database", key.Name())
		assert.Equal(t, "postgresql://localhost:5432", key.Value())
		assert.Equal(t, int64(3), key.Version())
		assert.Equal(t, int64(10), key.CreateRevision())
		assert.Equal(t, int64(12), key.ModRevision())
		assert.Equal(t, int64(42), key.Lease())
	})

	t.Run("holds an empty value and no lease when the store answered with none", func(t *testing.T) {
		t.Parallel()

		key := ability.NewKey(&mvccpb.KeyValue{Key: []byte("app/config/empty"), Version: 1})

		assert.True(t, key.Exists())
		assert.Empty(t, key.Value())
		assert.Zero(t, key.Lease())
	})

	t.Run("holds nothing but its name when it does not exist", func(t *testing.T) {
		t.Parallel()

		key := ability.NewMissingKey("does/not/exist")

		assert.False(t, key.Exists())
		assert.Equal(t, "does/not/exist", key.Name())
		assert.Empty(t, key.Value())
		assert.Zero(t, key.Version())
		assert.Zero(t, key.CreateRevision())
		assert.Zero(t, key.ModRevision())
		assert.Zero(t, key.Lease())
	})
}
