package resolution_test

import (
	"testing"

	"github.com/grandper/go-screenplay/v2/resolution/fixture"
	"github.com/stretchr/testify/assert"

	etcdfixture "github.com/grandper/go-screenplay-etcd/question/fixture"
	"github.com/grandper/go-screenplay-etcd/resolution"
)

func TestExist(t *testing.T) {
	existingKey := etcdfixture.CreateFakeExistingKey("app/status", "ready")
	missingKey := etcdfixture.CreateFakeMissingKey("does/not/exist")

	t.Run("matches a key that exists", func(t *testing.T) {
		fixture.AssertMatch(t, resolution.Exist(), existingKey)
	})

	t.Run("does not match a key that does not exist", func(t *testing.T) {
		fixture.AssertNoMatch(t, resolution.Exist(), missingKey)
	})

	t.Run("does not match a nil key", func(t *testing.T) {
		fixture.AssertNoMatch(t, resolution.Exist(), nil)
	})

	t.Run("implements the stringer interface", func(t *testing.T) {
		assert.Equal(t, "exist", resolution.Exist().String())
	})
}
