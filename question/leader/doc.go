// Package leader holds the question that answers with the value of the leader
// of an election:
//
//	err := anActor.Should(see.The(leader.OfTheElection("primary-worker")).Should(be.EqualTo("Alice")))
//
// reads "see the leader of the election primary-worker should be equal to Alice".
package leader
