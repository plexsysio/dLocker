package memlock

import (
	"context"
	"errors"
	"sync"
	"time"
)

const separator = "-"

var mtx sync.Mutex

type memLocker struct {
	smap map[string]*semaphore
}

// NewLocker return new lock
func NewLocker() *memLocker {
	return &memLocker{
		smap: make(map[string]*semaphore),
	}
}

func (m *memLocker) Close() error {
	m.smap = make(map[string]*semaphore)
	return nil
}

func (m *memLocker) getSemaphore(key string) *semaphore {
	mtx.Lock()
	defer mtx.Unlock()
	v, ok := m.smap[key]
	if !ok {
		m.smap[key] = newSemaphore()
		v = m.smap[key]
	}
	return v
}

func (m *memLocker) TryLock(
	ctx context.Context,
	key string,
	timeout time.Duration,
) (func(), error) {
	s := m.getSemaphore(key)

	if !s.tryLock(ctx, timeout) {
		return nil, errors.New("tryLock timeout")
	}
	return func() {
		s.unlock()
	}, nil

}

func newSemaphore() *semaphore {
	s := make(semaphore, 1)
	return &s
}

type semaphore chan struct{}

func (s semaphore) unlock() {
	<-s
}

func (s semaphore) tryLock(ctx context.Context, timeout time.Duration) bool {
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-ctx.Done():
	case s <- struct{}{}:
		return true
	case <-time.After(timeout):
	}
	return false
}
