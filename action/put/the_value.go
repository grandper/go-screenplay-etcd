package put

import impl "github.com/grandper/go-screenplay-etcd/internal/action/put"

// TheValue puts a value in the store. The key is given with ForTheKey, so the
// statement reads as a sentence:
//
//	put.TheValue("postgresql://localhost:5432").ForTheKey("app/config/database")
func TheValue(value string) *impl.TheValueAction {
	return impl.TheValue(value)
}
