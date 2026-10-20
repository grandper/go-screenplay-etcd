package fixture_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/grandper/go-screenplay-etcd/question/fixture"
)

func TestFakes(t *testing.T) {
	t.Run("creates a fake key that exists", func(t *testing.T) {
		key := fixture.CreateFakeExistingKey("app/status", "ready")

		assert.True(t, key.Exists())
		assert.Equal(t, "app/status", key.Name())
		assert.Equal(t, "ready", key.Value())
		assert.Equal(t, int64(1), key.Version())
		assert.Positive(t, key.CreateRevision())
		assert.Positive(t, key.ModRevision())
	})

	t.Run("creates a fake key that does not exist", func(t *testing.T) {
		key := fixture.CreateFakeMissingKey("does/not/exist")

		assert.False(t, key.Exists())
		assert.Equal(t, "does/not/exist", key.Name())
		assert.Empty(t, key.Value())
		assert.Zero(t, key.Version())
	})
}
