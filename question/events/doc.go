// Package events holds the questions that answer with the events recorded on
// a watched key, or on the watched keys starting with a prefix:
//
//	err := anActor.Should(see.The(last.Of(events.OnTheKey("app/status"))).ShouldHave(been.Put()))
//
// reads "see the last of events on the key app/status should have been put".
package events
