package campaign

import impl "github.com/grandper/go-screenplay-etcd/internal/action/campaign"

// InTheElection campaigns in an election, waiting for the leader to resign if
// there is one, and returns once the actor leads it. The leadership is held
// until the actor resigns with resign.FromTheElection, or exits. The leader
// is announced under the name of the actor, unless a value is given with
// WithTheValue:
//
//	campaign.InTheElection("primary-worker")
func InTheElection(name string) *impl.InTheElectionAction {
	return impl.InTheElection(name)
}
