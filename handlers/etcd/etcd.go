package etcd

import (
	"context"
	"time"

	logger "github.com/ipfs/go-log/v2"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

var log = logger.Logger("locker/etcd")

type etcdLocker struct {
	cli *clientv3.Client
}

func New(endpoint string) (*etcdLocker, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{endpoint},
	})
	if err != nil {
		return nil, err
	}

	return &etcdLocker{cli: cli}, nil
}

func (l *etcdLocker) Close() error {
	return l.cli.Close()
}

func (l *etcdLocker) TryLock(
	ctx context.Context,
	key string,
	timeout time.Duration,
) (func(), error) {

	s, err := concurrency.NewSession(l.cli)
	if err != nil {
		return func() {}, err
	}
	lk := concurrency.NewMutex(s, key)

	acquired := make(chan struct{})

	go func() {
		defer close(acquired)

		cctx, ccancel := context.WithTimeout(ctx, timeout)
		defer ccancel()

		err = lk.Lock(cctx)
	}()

	select {
	case <-acquired:
		if err == nil {
			log.Debugf("lock acquired %s", key)
			return func() {
				lk.Unlock(ctx)
				s.Close()
			}, nil
		}
	case <-ctx.Done():
		err = ctx.Err()
	}
	return func() {}, err
}
