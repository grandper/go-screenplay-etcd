package ability

import (
	"context"
	"sync"
)

// background holds the goroutines an ability runs on behalf of the actor,
// such as the renewals of a lease kept alive, so that they all stop when the
// ability is forgotten.
type background struct {
	mutex   sync.Mutex
	group   sync.WaitGroup
	cancels []context.CancelFunc
}

// run starts the function in a goroutine, with a context canceled when the
// background stops, and returns the function that cancels this goroutine
// alone, for the caller that gives up on it before the background stops.
func (b *background) run(run func(ctx context.Context)) context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())

	b.mutex.Lock()
	b.cancels = append(b.cancels, cancel)
	b.mutex.Unlock()

	b.group.Go(func() {
		run(ctx)
	})

	return cancel
}

// stop cancels every goroutine and waits for all of them to return.
func (b *background) stop() {
	b.mutex.Lock()
	cancels := b.cancels
	b.cancels = nil
	b.mutex.Unlock()

	for _, cancel := range cancels {
		cancel()
	}

	b.group.Wait()
}
