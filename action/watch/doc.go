// Package watch holds the actions that watch keys of the store, recording
// every change made to them until the actor exits:
//
//	err := anActor.AttemptsTo(watch.TheKey("app/config/database"))
//	err := anActor.AttemptsTo(watch.TheKeysStartingWith("app/config/"))
//
// The recorded changes are asked about with events.OnTheKey and
// events.OnTheKeysStartingWith.
package watch
