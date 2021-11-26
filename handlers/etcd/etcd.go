package etcd

import (
	"context"
	"errors"
	"time"

	logger "github.com/ipfs/go-log/v2"
	cpool "github.com/plexsysio/conn-pool"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"google.golang.org/grpc/connectivity"
)

var log = logger.Logger("locker/etcd")

type etcdLocker struct {
	pool *cpool.Pool
}

type etcdConn struct {
	cli  *clientv3.Client
	sess *concurrency.Session
}

func New(endpoint string) (*etcdLocker, error) {
	cpool, err := cpool.New(
		endpoint,
		func(ep string) (cpool.Conn, error) {
			cli, err := clientv3.New(clientv3.Config{
				Endpoints: []string{endpoint},
			})
			if err != nil {
				return nil, err
			}
			sess, err := concurrency.NewSession(cli)
			if err != nil {
				return nil, err
			}
			return &etcdConn{
				cli:  cli,
				sess: sess,
			}, nil
		},
		func(c cpool.Conn) error {
			if c.(*etcdConn).cli.ActiveConnection().GetState() == connectivity.Shutdown {
				return errors.New("invalid grpc conn state")
			}
			return nil
		},
		func(c cpool.Conn) error {
			c.(*etcdConn).sess.Close()
			c.(*etcdConn).cli.Close()
			return nil
		},
		100, 20, 15*time.Minute,
	)
	if err != nil {
		return nil, err
	}

	return &etcdLocker{pool: cpool}, nil
}

func (l *etcdLocker) Close() error {
	return l.pool.Close()
}

func (l *etcdLocker) TryLock(
	ctx context.Context,
	key string,
	timeout time.Duration,
) (func(), error) {

	econn, done, err := l.pool.GetContext(ctx)
	if err != nil {
		return nil, err
	}

	lk := concurrency.NewMutex(econn.(*etcdConn).sess, key)

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
				done()
			}, nil
		}
	case <-ctx.Done():
		err = ctx.Err()
	}
	return nil, err
}
