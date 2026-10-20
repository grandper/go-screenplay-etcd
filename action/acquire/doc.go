// Package acquire holds the action that acquires a distributed lock:
//
//	err := anActor.AttemptsTo(acquire.TheLock("deployment"))
//	err := anActor.AttemptsTo(acquire.TheLock("deployment").WithoutWaiting())
package acquire
