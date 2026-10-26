package exist

import (
	"errors"

	"github.com/grandper/go-screenplay/v2/resolution"
	"github.com/grandper/go-screenplay/v2/screenplay"

	"github.com/grandper/go-screenplay-etcd/etcd"
)

// InETCD creates a resolution to tell if a key exists in etcd, for the key
// answered by key.Named:
//
//	see.The(key.Named("app/status")).Should(exist.InETCD())
//
// A nil key does not exist.
func InETCD() screenplay.Resolution[*etcd.Key] {
	return resolution.FromFunc(
		"exist in etcd",
		func(key *etcd.Key) error {
			if key == nil || !key.Exists() {
				return errors.New("it does not exist")
			}

			return nil
		},
	)
}
