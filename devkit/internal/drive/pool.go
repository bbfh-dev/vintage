package drive

import (
	"sync"
)

const DEFAULT_POOL_SIZE = 5000

type Pool[T any] struct {
	mutex     sync.Mutex
	pool      []*T
	index     int
	chunkSize int
}

// Acquire returns a JsonFile with a copy of the given body.
// The pool grows automatically if more objects are needed.
func (pool *Pool[T]) Acquire(override func(*T)) *T {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	// Grow the pool if needed
	if pool.index >= len(pool.pool) {
		newPool := make([]*T, len(pool.pool)+pool.chunkSize)
		copy(newPool, pool.pool)
		for i := len(pool.pool); i < len(newPool); i++ {
			newPool[i] = new(T)
		}
		pool.pool = newPool
	}

	file := pool.pool[pool.index]
	pool.index++

	override(file)
	return file
}

func NewPool[T any](initialSize, chunkSize int) *Pool[T] {
	p := &Pool[T]{
		pool:      make([]*T, initialSize),
		index:     0,
		chunkSize: chunkSize,
	}
	for i := range initialSize {
		p.pool[i] = new(T)
	}
	return p
}
