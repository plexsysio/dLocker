package zookeeper

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-zookeeper/zk"
	logger "github.com/ipfs/go-log/v2"
	"github.com/plexsysio/conn-pool"
)

var log = logger.Logger("locker/zk")

type zkLocker struct {
	zp *cpool.Pool
}

func NewZkLocker(
	zkHost string,
	zkPort int,
) (*zkLocker, error) {
	p, err := cpool.New(
		fmt.Sprintf("%s:%d", zkHost, zkPort),
		func(addr string) (cpool.Conn, error) {
			log.Debugf("connecting to %s", addr)
			c, _, err := zk.Connect([]string{addr}, time.Millisecond*100)
			if err != nil {
				log.Errorf("failed connecting to zookeeper Err:%s", err.Error())
				return nil, err
			}
			return c, err
		},
		func(conn cpool.Conn) error {
			if zkConnObj, ok := conn.(*zk.Conn); ok {
				if zkConnObj.State() == zk.StateDisconnected ||
					zkConnObj.State() == zk.StateExpired {
					return errors.New("connection state wrong")
				}
			}
			return nil
		},
		func(conn cpool.Conn) error {
			if zkConnObj, ok := conn.(*zk.Conn); ok {
				zkConnObj.Close()
			}
			return nil
		},
		50, 10, time.Minute*15,
	)
	if err != nil {
		log.Errorf("failed creating new connection pool %s", err.Error())
		return nil, err
	}
	return &zkLocker{zp: p}, nil
}

func (l *zkLocker) Close() error {
	return l.zp.Close()
}

func (l *zkLocker) TryLock(
	ctx context.Context,
	key string,
	t time.Duration,
) (func(), error) {
	conn, done, e := l.zp.GetContext(ctx)
	if e != nil {
		log.Errorf("failed creating connection %s", e.Error())
		return nil, e
	}

	if !strings.HasPrefix(key, "/") {
		key = "/" + key
	}
	zl := zk.NewLock(conn.(*zk.Conn), key, zk.WorldACL(zk.PermAll))

	ch := make(chan error)
	stopped := make(chan struct{})
	defer close(stopped)

	go func() {
		err := zl.Lock()
		select {
		case <-stopped:
			log.Info("lock attempt was timed out, giving up")
			zl.Unlock()
			return
		case ch <- err:
		}
	}()

	var err error
	select {
	case err = <-ch:
		if err == nil {
			log.Debugf("lock acquired %s", key)
			return func() {
				zl.Unlock()
				done()
			}, nil
		}
	case <-time.After(t):
		err = errors.New("timed out getting lock")
	case <-ctx.Done():
		err = ctx.Err()
	}
	log.Errorf("failed getting lock Err %s", err.Error())
	return nil, err
}
