package been_test

import (
	"testing"

	"github.com/grandper/go-screenplay/v2/resolution/fixture"
	"github.com/stretchr/testify/assert"

	etcdfixture "github.com/grandper/go-screenplay-etcd/internal/fixture"
	"github.com/grandper/go-screenplay-etcd/resolution/been"
)

func TestPut(t *testing.T) {
	t.Run("matches a put event", func(t *testing.T) {
		fixture.AssertMatch(t, been.Put(), etcdfixture.CreateFakePutEvent("app/status", "ready"))
	})

	t.Run("does not match a delete event", func(t *testing.T) {
		fixture.AssertNoMatch(t, been.Put(), etcdfixture.CreateFakeDeleteEvent("app/status", "ready"))
		fixture.AssertNoMatch(t, been.Put(), etcdfixture.CreateFakeExpiryEvent("app/status", "ready"))
	})

	t.Run("does not match a nil event", func(t *testing.T) {
		fixture.AssertNoMatch(t, been.Put(), nil)
	})

	t.Run("implements the stringer interface", func(t *testing.T) {
		assert.Equal(t, "been put", been.Put().String())
	})
}
