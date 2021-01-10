package dLocker

import (
	"context"
	"time"
)

type DLocker interface {
	Close() error
	// Try to obtain lock for 'duration' then fail
	TryLock(context.Context, string, time.Duration) (func(), error)
}
