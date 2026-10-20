// Package been holds the resolutions about an event recorded on a watched
// key. They are past participles and take the verb ShouldHave (or
// ShouldNotHave):
//
//	err := anActor.Should(see.The(last.Of(question.EventsOnTheKey("app/status"))).ShouldHave(been.Put()))
//
// reads "see the last of events on the key app/status should have been put".
package been
