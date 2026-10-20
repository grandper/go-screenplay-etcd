// Package remove holds the actions that remove keys from the store:
//
//	err := anActor.AttemptsTo(remove.TheKey("app/config/cache"))
//	err := anActor.AttemptsTo(remove.TheKeys("app/config/cache", "app/config/database"))
//	err := anActor.AttemptsTo(remove.TheKeysStartingWith("app/config/"))
package remove
