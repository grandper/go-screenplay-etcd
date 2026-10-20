package resolution

import (
	"errors"

	"github.com/grandper/go-screenplay/v2/resolution"
	"github.com/grandper/go-screenplay/v2/screenplay"

	"github.com/grandper/go-screenplay-etcd/ability"
)

// Exist creates a resolution to tell if a key exists, for the key answered by
// question.KeyNamed:
//
//	see.The(question.KeyNamed("app/status")).Should(resolution.Exist())
//
// A nil key does not exist.
func Exist() screenplay.Resolution[*ability.Key] {
	return resolution.FromFunc(
		"exist",
		func(key *ability.Key) error {
			if key == nil || !key.Exists() {
				return errors.New("it does not exist")
			}

			return nil
		},
	)
}
