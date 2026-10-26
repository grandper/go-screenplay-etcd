package etcd

import impl "github.com/grandper/go-screenplay-etcd/internal/ability"

// Key is a key of the store as the actor saw it. Its parts are read through
// Name(), Exists(), Value(), Version(), CreateRevision(), ModRevision() and
// Lease().
type Key = impl.Key
