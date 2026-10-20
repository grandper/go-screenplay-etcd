// Package revision holds the question that answers with the revision of the
// last modification of a key:
//
//	revision, err := anActor.AsksFor(revision.OfTheKey("app/config/database")) // revision is an int64
//
// reads "asks for the revision of the key app/config/database".
package revision
