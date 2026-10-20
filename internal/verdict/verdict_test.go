package verdict_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/grandper/go-screenplay-etcd/internal/verdict"
)

func TestVerdict(t *testing.T) {
	t.Parallel()

	t.Run("describe", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, `"hello"`, verdict.Describe("hello"))
		assert.Equal(t, `"hello"`, verdict.Describe([]byte("hello")))
		assert.Equal(t, "42", verdict.Describe(42))
		assert.Equal(t, "[a b]", verdict.Describe([]string{"a", "b"}))
		assert.Equal(t, "<nil>", verdict.Describe(nil))
	})
}
