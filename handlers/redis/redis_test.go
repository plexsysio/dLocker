package redis

import (
	"github.com/aloknerurkar/dLocker/testsuite"
	logger "github.com/ipfs/go-log/v2"
	"github.com/stvp/tempredis"
	"testing"
)

func TestLockerSuite(t *testing.T) {
	logger.SetLogLevel("locker/redis", "Debug")
	srv, err := tempredis.Start(tempredis.Config{})
	if err != nil {
		panic(err)
	}
	defer srv.Term()
	testsuite.RunTests(t, func() interface{} {
		return NewRedisLocker("unix", srv.Socket())
	})
}
