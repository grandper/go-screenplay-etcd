// Package values holds the question that answers with the values of the keys
// starting with a prefix, by key:
//
//	err := anActor.Should(see.The(values.OfTheKeysStartingWith("app/config/")).Should(contain.Key[string, string]("app/config/database")))
//
// reads "see the values of the keys starting with app/config/ should contain key app/config/database".
package values
