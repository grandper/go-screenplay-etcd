package acquire

import impl "github.com/grandper/go-screenplay-etcd/action/internal/acquire"

// TheLock acquires a distributed lock, waiting for the actor holding it to
// release it. The lock is held until the actor releases it with
// release.TheLock, or exits:
//
//	acquire.TheLock("deployment")
func TheLock(name string) *impl.TheLockAction {
	return impl.TheLock(name)
}
