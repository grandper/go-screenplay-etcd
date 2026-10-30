// Package put holds the actions that put values in the store:
//
//	err := anActor.AttemptsTo(put.TheValue("postgresql://localhost:5432").ForTheKey("app/config/database"))
//
// reads "put the value postgresql://localhost:5432 for the key app/config/database".
// Several values are put in one transaction with TheValues:
//
//	err := anActor.AttemptsTo(put.TheValues(map[string]string{
//		"app/config/database": "postgresql://localhost:5432",
//		"app/config/cache":    "redis://localhost:6379",
//	}))
package put
