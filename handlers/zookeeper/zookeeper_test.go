package zookeeper_test

import (
	"os"
	"strconv"
	"testing"

	logger "github.com/ipfs/go-log/v2"
	"github.com/plexsysio/dLocker/handlers/zookeeper"
	"github.com/plexsysio/dLocker/testsuite"
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
	l, err := zookeeper.NewZkLocker(zkHostname(), zkPort())
	if err != nil {
		t.Fatal("unable to connect to zookeeper", err)
	}
	t.Cleanup(func() {
		err := l.Close()
		if err != nil {
			t.Fatal(err)
		}
	})
	testsuite.RunTests(t, l)
}

func BenchmarkSuite(b *testing.B) {
	logger.SetLogLevel("locker/zk", "Error")
	l, err := zookeeper.NewZkLocker(zkHostname(), zkPort())
	if err != nil {
		b.Fatal("unable to connect to zookeeper", err)
	}
	b.Cleanup(func() {
		err := l.Close()
		if err != nil {
			b.Fatal(err)
		}
	})
	testsuite.RunBenchmark(b, l)
}
