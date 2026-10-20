// Package resolution holds the resolution to tell if a key exists. It is a
// bare infinitive and takes the verb Should (or ShouldNot):
//
//	err := anActor.Should(see.The(question.KeyNamed("app/status")).Should(resolution.Exist()))
//
// reads "see the key named app/status should exist".
package resolution
