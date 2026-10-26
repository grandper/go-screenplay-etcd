package etcd_test

import (
	"testing"
	"time"

	"github.com/grandper/go-screenplay/v2/action/wait"
	"github.com/grandper/go-screenplay/v2/resolution/length"
	"github.com/grandper/go-screenplay/v2/screenplay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grandper/go-screenplay-etcd/action/put"
	"github.com/grandper/go-screenplay-etcd/action/remove"
	"github.com/grandper/go-screenplay-etcd/action/watch"
	"github.com/grandper/go-screenplay-etcd/etcd"
	"github.com/grandper/go-screenplay-etcd/fixture"
	"github.com/grandper/go-screenplay-etcd/question/events"
	"github.com/grandper/go-screenplay-etcd/use"
)

func TestEvent(t *testing.T) {
	server := fixture.RunETCDServer(t)
	production := screenplay.NewProduction(screenplay.WithTimeout(15 * time.Second))
	alice := production.SetTheStage(screenplay.CastOfStandardActors()).
		ActorNamed("Alice").
		WhoCan(use.ETCD().WithTheEndpoints(server.Endpoint()))
	t.Cleanup(func() { require.NoError(t, alice.Exit()) })

	require.NoError(t, alice.AttemptsTo(
		watch.TheKey("app/status"),
		put.TheValue("ready").ForTheKey("app/status"),
		put.TheValue("stopped").ForTheKey("app/status"),
		remove.TheKey("app/status"),
		put.TheValue("starting").ForTheKey("app/status").ExpiringAfter(1).Second(),
		wait.UntilThe(events.OnTheKey("app/status")).HappensToHave(length.Exactly[[]*etcd.Event](5)),
	))
	theEvents, err := alice.AsksFor(events.OnTheKey("app/status"))
	require.NoError(t, err)
	require.Len(t, theEvents, 5)
	created, overwritten, removed, expired := theEvents[0], theEvents[1], theEvents[2], theEvents[4]

	t.Run("exposes the parts of a put that created the key", func(t *testing.T) {
		assert.True(t, created.IsPut())
		assert.False(t, created.IsDelete())
		assert.Equal(t, "app/status", created.Key())
		assert.Equal(t, "ready", created.Value())
		assert.Empty(t, created.PreviousValue())
		assert.Positive(t, created.Revision())
		assert.Equal(t, int64(1), created.Version())
		assert.False(t, created.CausedByLeaseExpiry())
	})

	t.Run("exposes the previous value of a put that overwrote the key", func(t *testing.T) {
		assert.True(t, overwritten.IsPut())
		assert.Equal(t, "stopped", overwritten.Value())
		assert.Equal(t, "ready", overwritten.PreviousValue())
		assert.Greater(t, overwritten.Revision(), created.Revision())
		assert.Equal(t, int64(2), overwritten.Version())
	})

	t.Run("exposes the parts of a delete by an actor", func(t *testing.T) {
		assert.True(t, removed.IsDelete())
		assert.False(t, removed.IsPut())
		assert.Equal(t, "app/status", removed.Key())
		assert.Empty(t, removed.Value())
		assert.Equal(t, "stopped", removed.PreviousValue())
		assert.Greater(t, removed.Revision(), overwritten.Revision())
		assert.Zero(t, removed.Version())
		assert.False(t, removed.CausedByLeaseExpiry())
	})

	t.Run("exposes a delete caused by the expiry of the lease", func(t *testing.T) {
		assert.True(t, expired.IsDelete())
		assert.Equal(t, "starting", expired.PreviousValue())
		assert.True(t, expired.CausedByLeaseExpiry())
	})
}
