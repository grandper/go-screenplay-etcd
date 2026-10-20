package been_test

import (
	"testing"

	"github.com/grandper/go-screenplay/v2/resolution/fixture"
	"github.com/stretchr/testify/assert"

	etcdfixture "github.com/grandper/go-screenplay-etcd/question/fixture"
	"github.com/grandper/go-screenplay-etcd/resolution/been"
)

func TestDeleted(t *testing.T) {
	putEvent := etcdfixture.CreateFakePutEvent("app/status", "ready")
	deleteEvent := etcdfixture.CreateFakeDeleteEvent("app/status", "ready")
	expiryEvent := etcdfixture.CreateFakeExpiryEvent("app/status", "ready")

	t.Run("matches a delete event, whatever its cause", func(t *testing.T) {
		fixture.AssertMatch(t, been.Deleted(), deleteEvent)
		fixture.AssertMatch(t, been.Deleted(), expiryEvent)
	})

	t.Run("does not match a put event", func(t *testing.T) {
		fixture.AssertNoMatch(t, been.Deleted(), putEvent)
	})

	t.Run("does not match a nil event", func(t *testing.T) {
		fixture.AssertNoMatch(t, been.Deleted(), nil)
		fixture.AssertNoMatch(t, been.Deleted().ByTheExpiryOfItsLease(), nil)
	})

	t.Run("matches only a delete event caused by the expiry of a lease when asked to", func(t *testing.T) {
		fixture.AssertMatch(t, been.Deleted().ByTheExpiryOfItsLease(), expiryEvent)
		fixture.AssertNoMatch(t, been.Deleted().ByTheExpiryOfItsLease(), deleteEvent)
		fixture.AssertNoMatch(t, been.Deleted().ByTheExpiryOfItsLease(), putEvent)
	})

	t.Run("implements the stringer interface", func(t *testing.T) {
		assert.Equal(t, "been deleted", been.Deleted().String())
		assert.Equal(t, "been deleted by the expiry of its lease", been.Deleted().ByTheExpiryOfItsLease().String())
	})
}
