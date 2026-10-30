// Package been holds the resolutions about an event recorded on a watched
// key. They are past participles and take the verb ShouldHave (or
// ShouldNotHave):
//
//	err := anActor.Should(see.The(last.Of(events.OnTheKey("app/status"))).ShouldHave(been.Put()))
//
// reads "see the last of events on the key app/status should have been put".
// A delete is told from a put with Deleted, closed with ByTheExpiryOfItsLease
// when the key must have expired rather than been removed by an actor:
//
//	err := anActor.Should(see.The(last.Of(events.OnTheKey("locks/batch-job"))).ShouldHave(been.Deleted().ByTheExpiryOfItsLease()))
package been
