package fixture

import (
	"go.etcd.io/etcd/api/v3/mvccpb"

	"github.com/grandper/go-screenplay-etcd/ability"
	impl "github.com/grandper/go-screenplay-etcd/internal/ability"
)

// CreateFakeExistingKey creates a fake key that exists, holding the value, as
// if it had just been created at the first revision of the store.
func CreateFakeExistingKey(name, value string) *ability.Key {
	return impl.NewKey(&mvccpb.KeyValue{
		Key:            []byte(name),
		Value:          []byte(value),
		Version:        1,
		CreateRevision: 1,
		ModRevision:    1,
	})
}

// CreateFakeMissingKey creates a fake key that does not exist.
func CreateFakeMissingKey(name string) *ability.Key {
	return impl.NewMissingKey(name)
}
