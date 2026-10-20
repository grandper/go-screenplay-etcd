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

func TestFakeEvents(t *testing.T) {
	t.Run("creates a fake put event", func(t *testing.T) {
		event := fixture.CreateFakePutEvent("app/status", "ready")

		assert.True(t, event.IsPut())
		assert.False(t, event.IsDelete())
		assert.Equal(t, "app/status", event.Key())
		assert.Equal(t, "ready", event.Value())
		assert.Empty(t, event.PreviousValue())
		assert.Positive(t, event.Revision())
		assert.Equal(t, int64(1), event.Version())
		assert.False(t, event.CausedByLeaseExpiry())
	})

	t.Run("creates a fake delete event", func(t *testing.T) {
		event := fixture.CreateFakeDeleteEvent("app/status", "ready")

		assert.True(t, event.IsDelete())
		assert.False(t, event.IsPut())
		assert.Equal(t, "app/status", event.Key())
		assert.Empty(t, event.Value())
		assert.Equal(t, "ready", event.PreviousValue())
		assert.False(t, event.CausedByLeaseExpiry())
	})

	t.Run("creates a fake delete event caused by the expiry of a lease", func(t *testing.T) {
		event := fixture.CreateFakeExpiryEvent("app/status", "ready")

		assert.True(t, event.IsDelete())
		assert.Equal(t, "ready", event.PreviousValue())
		assert.True(t, event.CausedByLeaseExpiry())
	})
}
