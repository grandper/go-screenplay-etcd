package remove

import impl "github.com/grandper/go-screenplay-etcd/internal/action/remove"

// TheKey removes a key from the store. Removing a key that does not exist is
// not an error:
//
//	remove.TheKey("app/config/cache")
func TheKey(key string) *impl.TheKeyAction {
	return impl.TheKey(key)
}
