package redis

import (
	"context"
	"errors"
	"time"

	goredislib "github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
	logger "github.com/ipfs/go-log/v2"
)

var log = logger.Logger("locker/redis")

type redisLocker struct {
	rs      *redsync.Redsync
	closeFn func() error
}

var (
	// ErrTimeout returns when you couldn't make a TryLock call
	ErrTimeout = errors.New("timeout reached")

	// ErrCancelled returns when you have cancelled the context of req
	ErrCancelled = errors.New("context cancelled")

	// Default lock timeout
	DefaultTimeout time.Duration = time.Second * 8
	DefaultTicker  time.Duration = time.Second * 7
)

type Option func(*goredislib.Options)

func WithUsernamePassword(username, password string) Option {
	return func(opts *goredislib.Options) {
		opts.Username = username
		opts.Password = password
	}
}

func NewRedisLocker(netw, server string, opts ...Option) *redisLocker {
	rdOpts := &goredislib.Options{
		Network: netw,
		Addr:    server,
	}
	for _, opt := range opts {
		opt(rdOpts)
	}
	client := goredislib.NewClient(rdOpts)
	pool := goredis.NewPool(client)
	return &redisLocker{
		rs: redsync.New(pool),
		closeFn: func() error {
			return client.Close()
		},
	}
}

func (l *redisLocker) Close() error {
	return l.closeFn()
}

func (l *redisLocker) TryLock(
	ctx context.Context,
	key string,
	timeout time.Duration,
) (func(), error) {

	mtx := l.rs.NewMutex(key, redsync.WithExpiry(DefaultTimeout))

	acquired := make(chan struct{})
	var err error

	go func() {
		defer close(acquired)

		cCtx, cCancel := context.WithTimeout(ctx, timeout)
		defer cCancel()

		err = mtx.LockContext(cCtx)
	}()

	select {
	case <-acquired:
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				return nil, ErrTimeout
			}
			return nil, err
		}
		stop := make(chan struct{})
		go func() {
			t := time.NewTicker(DefaultTicker)
			for {
				select {
				case <-stop:
					log.Info("lock extender stopped")
				case <-t.C:
					v, err := mtx.ValidContext(ctx)
					if err != nil {
						log.Errorf("failed checking Mutex validity Err:%s", err.Error())
						return
					}
					if !v {
						log.Errorf("mutex no longer valid")
						return
					}
					extended, err := mtx.ExtendContext(ctx)
					if err != nil {
						log.Errorf("failed extending lock Err:%s", err.Error())
						return
					}
					if !extended {
						log.Errorf("unable to extend mutex")
						return
					}
					log.Infof("lock extended")
					continue
				}
				return
			}
		}()
		log.Debugf("lock acquired %s", key)
		return func() {
			close(stop)
			mtx.UnlockContext(ctx)
		}, nil
	case <-ctx.Done():
		return nil, ErrCancelled
	}
}
