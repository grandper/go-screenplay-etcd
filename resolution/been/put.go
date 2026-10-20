package been

import (
	"errors"

	"github.com/grandper/go-screenplay/v2/resolution"
	"github.com/grandper/go-screenplay/v2/screenplay"

	"github.com/grandper/go-screenplay-etcd/ability"
)

// Put creates a resolution to tell if an event put a value for its key:
//
//	see.The(last.Of(question.EventsOnTheKey("app/status"))).ShouldHave(been.Put())
//
// A nil event is not a put.
func Put() screenplay.Resolution[*ability.Event] {
	return resolution.FromFunc(
		"been put",
		func(event *ability.Event) error {
			if event == nil {
				return errors.New("there is no event")
			}

			if !event.IsPut() {
				return errors.New("it has been deleted")
			}

			return nil
		},
	)
}
