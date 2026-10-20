// Package put holds the actions that put values in the store:
//
//	err := anActor.AttemptsTo(put.TheValue("postgresql://localhost:5432").ForTheKey("app/config/database"))
//
// reads "put the value postgresql://localhost:5432 for the key app/config/database".
package put
