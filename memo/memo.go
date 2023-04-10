package memo

import (
	"context"
	"sync"

	"github.com/go-logr/logr"
)

// Func is the type of the function to memoize.
type Func[K comparable, V any] func(ctx context.Context, key K, cleanup func()) (V, error)

type result[V any] struct {
	value V
	err   error
}

type entry[V any] struct {
	res   result[V]
	ready chan struct{} // closed when res is ready
}

func New[K comparable, V any](f Func[K, V]) *Memo[K, V] {
	return &Memo[K, V]{Log: logr.Discard(), f: f, cache: make(map[K]*entry[V])}
}

type Memo[K comparable, V any] struct {
	// Log specifies an optional logger for cleanup.
	// If empty, logging is discard
	Log logr.Logger

	f     Func[K, V]
	mu    sync.Mutex // guards cache
	cache map[K]*entry[V]
}

func (memo *Memo[K, V]) Get(ctx context.Context, key K) (value V, err error) {
	memo.mu.Lock()
	e := memo.cache[key]
	if e == nil {
		// This is the first request for this key.
		// This goroutine becomes responsible for computing
		// the value and broadcasting the ready condition.
		e = &entry[V]{ready: make(chan struct{})}
		memo.cache[key] = e
		memo.mu.Unlock()

		e.res.value, e.res.err = memo.f(ctx, key, func() {
			memo.remove(key)
		})

		close(e.ready) // broadcast ready condition
	} else {
		// This is a repeat request for this key.
		memo.mu.Unlock()

		<-e.ready // wait for ready condition
	}
	return e.res.value, e.res.err
}

func (memo *Memo[K, V]) remove(key K) {
	memo.mu.Lock()
	delete(memo.cache, key)
	memo.mu.Unlock()
}
