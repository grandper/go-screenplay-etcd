package put

import impl "github.com/grandper/go-screenplay-etcd/internal/action/put"

// TheValues puts several values in the store, each for its key, in one
// transaction, so that the store never holds some of them without the others:
//
//	put.TheValues(map[string]string{
//		"app/config/database": "postgresql://localhost:5432",
//		"app/config/cache":    "redis://localhost:6379",
//	})
func TheValues(values map[string]string) *impl.TheValuesAction {
	return impl.TheValues(values)
}
