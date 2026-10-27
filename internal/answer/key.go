package answer

import "go.etcd.io/etcd/api/v3/mvccpb"

// Key is a key of the store as the actor saw it: its value and its metadata
// when it exists, or just its name when it does not.
type Key struct {
	name           string
	value          string
	exists         bool
	version        int64
	createRevision int64
	modRevision    int64
	lease          int64
}

// NewKey creates the key the store answered with.
func NewKey(kv *mvccpb.KeyValue) *Key {
	return &Key{
		name:           string(kv.GetKey()),
		value:          string(kv.GetValue()),
		exists:         true,
		version:        kv.GetVersion(),
		createRevision: kv.GetCreateRevision(),
		modRevision:    kv.GetModRevision(),
		lease:          kv.GetLease(),
	}
}

// NewMissingKey creates a key that does not exist in the store.
func NewMissingKey(name string) *Key {
	return &Key{name: name}
}

// Name returns the name of the key.
func (k *Key) Name() string {
	return k.name
}

// Exists returns true when the key exists in the store.
func (k *Key) Exists() bool {
	return k.exists
}

// Value returns the value the key holds, empty when the key does not exist.
func (k *Key) Value() string {
	return k.value
}

// Version returns the number of times the key was put since it was created;
// it is zero when the key does not exist.
func (k *Key) Version() int64 {
	return k.version
}

// CreateRevision returns the revision of the store at which the key was
// created; it is zero when the key does not exist.
func (k *Key) CreateRevision() int64 {
	return k.createRevision
}

// ModRevision returns the revision of the store at which the key was last
// modified; it is zero when the key does not exist.
func (k *Key) ModRevision() int64 {
	return k.modRevision
}

// Lease returns the ID of the lease attached to the key, zero when the key has
// no lease or does not exist.
func (k *Key) Lease() int64 {
	return k.lease
}
