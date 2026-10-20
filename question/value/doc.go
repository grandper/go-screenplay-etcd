// Package value holds the question that answers with the value a key holds:
//
//	err := anActor.Should(see.The(value.OfTheKey("app/config/database")).Should(be.EqualTo("postgresql://localhost:5432")))
//
// reads "see the value of the key app/config/database should be equal to postgresql://localhost:5432".
package value
