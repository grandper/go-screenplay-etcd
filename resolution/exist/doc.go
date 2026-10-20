// Package exist holds the resolution to tell if a key exists. It is a bare
// infinitive and takes the verb Should (or ShouldNot):
//
//	err := anActor.Should(see.The(key.Named("app/status")).Should(exist.InETCD()))
//
// reads "see the key named app/status should exist in etcd".
package exist
