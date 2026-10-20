package question

import "errors"

// ErrKeyNotFound is wrapped by the error a question about one key returns
// when the key does not exist in the store.
var ErrKeyNotFound = errors.New("was not found")
