// Package ttl holds the question that answers with the time to live of a key,
// the time left before its lease expires:
//
//	err := anActor.Should(see.The(ttl.OfTheKey("locks/batch-job")).Should(be.AtMost(10*time.Second)))
//
// reads "see the TTL of the key locks/batch-job should be at most 10 seconds".
package ttl
