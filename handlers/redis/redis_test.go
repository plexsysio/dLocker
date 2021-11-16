package redis_test

import (
	"testing"

	logger "github.com/ipfs/go-log/v2"
	"github.com/plexsysio/dLocker/handlers/redis"
	"github.com/plexsysio/dLocker/testsuite"
	"github.com/stvp/tempredis"
)

func TestLockerSuite(t *testing.T) {
	logger.SetLogLevel("locker/redis", "Debug")
	srv, err := tempredis.Start(tempredis.Config{})
	if err != nil {
		t.Fatal("failed starting redis", err)
	}
	l := redis.NewRedisLocker("unix", srv.Socket())
	defer func() {
		err := l.Close()
		if err != nil {
			t.Fatal("failed closing redis locker", err)
		}
		srv.Term()
	}()
	testsuite.RunTests(t, l)
}

func BenchmarkSuite(b *testing.B) {
	logger.SetLogLevel("locker/redis", "Error")
	l := redis.NewRedisLocker("tcp", "localhost:6379")
	defer func() {
		err := l.Close()
		if err != nil {
			b.Fatal("failed closing redis locker", err)
		}
	}()
	testsuite.RunBenchmark(b, l)
}
