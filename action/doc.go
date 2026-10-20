// Package action holds the errors of the actions that drive an etcd store. The
// actions live in packages named after their verb:
//
//	put.TheValue("postgresql://localhost:5432").ForTheKey("app/config/database")  action/put
//
// For example:
//
//	err := anActor.AttemptsTo(
//		put.TheValue("postgresql://localhost:5432").ForTheKey("app/config/database"),
//	)
package action
