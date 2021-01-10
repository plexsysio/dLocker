package redis

import (
	"context"
	"errors"
	goredislib "github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
	logger "github.com/ipfs/go-log/v2"
	"time"
)

var log = logger.Logger("locker/redis")

type redisLocker struct {
	rs      *redsync.Redsync
	closeFn func() error
}

var (
	// ErrTimeout returns when you couldn't make a TryLock call
	ErrTimeout = errors.New("Timeout reached")

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
	t time.Duration,
) (func(), error) {
	ch := make(chan error)
	mtx := l.rs.NewMutex(key, redsync.WithExpiry(DefaultTimeout))

	cCtx, cancel := context.WithCancel(ctx)

	go func() {
		err := mtx.LockContext(cCtx)
		select {
		case ch <- err:
		default:
		}
	}()

	select {
	case err := <-ch:
		if err != nil {
			return nil, err
		}
		stop := make(chan bool)
		go func() {
			t := time.NewTicker(DefaultTicker)
			for {
				select {
				case <-cCtx.Done():
					log.Errorf("Context cancelled")
				case <-stop:
					log.Info("Lock extender stopped")
				case <-t.C:
					v, err := mtx.ValidContext(cCtx)
					if err != nil {
						log.Errorf("Failed checking Mutex validity Err:%s", err.Error())
						return
					}
					if !v {
						log.Errorf("Mutex no longer valid")
						return
					}
					extended, err := mtx.ExtendContext(cCtx)
					if err != nil {
						log.Errorf("Failed extending lock Err:%s", err.Error())
						return
					}
					if !extended {
						log.Errorf("Unable to extend mutex")
						return
					}
					log.Infof("Lock extended")
					continue
				}
				return
			}
		}()
		log.Debugf("Lock acquired %s", key)
		return func() {
			stop <- true
			mtx.UnlockContext(cCtx)
		}, nil
	case <-time.After(t):
		cancel()
		return nil, ErrTimeout
	case <-ctx.Done():
		return nil, ErrCancelled
	}
}
