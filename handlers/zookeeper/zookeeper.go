package zookeeper

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/plexsysio/conn-pool"
	"github.com/go-zookeeper/zk"
	logger "github.com/ipfs/go-log/v2"
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
			log.Debugf("Connecting to %s", addr)
			c, _, err := zk.Connect([]string{addr}, time.Millisecond*100)
			if err != nil {
				log.Errorf("Failed connecting to zookeeper Err:%s", err.Error())
				return nil, err
			}
			return c, err
		},
		func(conn cpool.Conn) error {
			if zkConnObj, ok := conn.(*zk.Conn); ok {
				if zkConnObj.State() == zk.StateDisconnected ||
					zkConnObj.State() == zk.StateExpired {
					return errors.New("Connection state wrong")
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
		log.Errorf("Failed creating new connection pool %s", err.Error())
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
	ch := make(chan error)
	conn, done, e := l.zp.GetContext(ctx)
	if e != nil {
		log.Errorf("Failed creating context based connection %s", e.Error())
		return nil, e
	}
	// We should be able to reuse the connection although the lock is dependent on it
	// as long as its not closed
	defer done()

	if !strings.HasPrefix(key, "/") {
		key = "/" + key
	}
	zl := zk.NewLock(conn.(*zk.Conn), key, zk.WorldACL(zk.PermAll))

	go func() {
		err := zl.Lock()
		log.Infof("Lock returned Err:%v", err)
		select {
		case _, open := <-ch:
			if !open {
				log.Info("Lock attempt was timed out, giving up")
				zl.Unlock()
				return
			}
		default:
		}
		select {
		case ch <- err:
		default:
		}
	}()

	select {
	case err := <-ch:
		if err != nil {
			log.Errorf("Failed getting lock Err %s", err.Error())
			return nil, err
		}
	case <-time.After(t):
		close(ch)
		return nil, errors.New("Timeout getting lock")
	case <-ctx.Done():
		return nil, errors.New("context cancelled")
	}
	log.Debugf("Lock acquired %s", key)
	return func() { zl.Unlock() }, nil
}
