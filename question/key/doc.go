// Package key holds the question that answers with a key, whether it exists
// or not:
//
//	err := anActor.Should(see.The(key.Named("app/status")).Should(exist.InETCD()))
//
// reads "see the key named app/status should exist in etcd".
package key
