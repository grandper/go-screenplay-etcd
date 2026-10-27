// Package fixture holds what the tests of the extension share: an embedded
// etcd server, plain or secured with TLS, an actor whose narration is
// recorded, helpers that read the store directly, and fakes of the answers
// of the questions for the tests of the resolutions:
//
//	server := fixture.RunETCDServer(t)
//	key := fixture.CreateFakeExistingKey("app/status", "ready")
package fixture
