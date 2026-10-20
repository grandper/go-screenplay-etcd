// Package question holds the errors shared by the questions of the extension.
// The questions themselves live in the packages named after their first word,
// so that each statement reads as English:
//
//	value, err := anActor.AsksFor(value.OfTheKey("app/config/database")) // value is a string
//	err := anActor.Should(see.The(key.Named("app/status")).Should(exist.InETCD()))
package question
