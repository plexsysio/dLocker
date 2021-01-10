package zookeeper

import (
	"github.com/aloknerurkar/dLocker/testsuite"
	logger "github.com/ipfs/go-log/v2"
	"os"
	"strconv"
	"testing"
)

var zkHostname = func() string {
	if len(os.Getenv("ZOOKEEPER_HOST")) > 0 {
		return os.Getenv("ZOOKEEPER_HOST")
	}
	return "localhost"
}

var zkPort = func() int {
	if len(os.Getenv("ZOOKEEPER_PORT")) > 0 {
		i, err := strconv.Atoi(os.Getenv("ZOOKEEPER_PORT"))
		if err == nil {
			return i
		}
	}
	return 2181
}

func TestLockerSuite(t *testing.T) {
	logger.SetLogLevel("locker/zk", "Debug")
	t.Log(zkHostname(), zkPort())
	testsuite.RunTests(t, func() interface{} {
		l, err := NewZkLocker(zkHostname(), zkPort())
		if err != nil {
			panic(err)
		}
		return l
	})
}
