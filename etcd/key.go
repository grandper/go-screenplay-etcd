package etcd

import "github.com/grandper/go-screenplay-etcd/internal/answer"

// Key is a key of the store as the actor saw it. Its parts are read through
// Name(), Exists(), Value(), Version(), CreateRevision(), ModRevision() and
// Lease().
type Key = answer.Key
