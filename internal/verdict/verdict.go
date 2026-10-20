// Package verdict holds the helper the actions and the questions of the
// extension share to describe the values they narrate.
package verdict

import "fmt"

// Describe renders a value as it reads in a description or in a failure:
// texts and bytes are quoted, everything else uses its default format.
func Describe(value any) string {
	switch v := value.(type) {
	case string:
		return fmt.Sprintf("%q", v)
	case []byte:
		return fmt.Sprintf("%q", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}
