// Package keys holds the question that answers with the names of the keys
// starting with a prefix:
//
//	err := anActor.Should(see.The(keys.StartingWith("app/config/")).ShouldHave(length.Exactly[[]string](2)))
//
// reads "see the keys starting with app/config/ should have length exactly 2".
package keys
