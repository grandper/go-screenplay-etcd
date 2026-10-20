package action

import "errors"

// ErrKeyNotProvided is returned when an action is performed without a key, or
// with an empty one. Nothing panics when the action is built; the mistake is
// reported when the actor performs it.
var ErrKeyNotProvided = errors.New("the key must be provided")
