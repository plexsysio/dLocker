package etcd_test

import (
	"testing"

	logger "github.com/ipfs/go-log/v2"
	"github.com/plexsysio/dLocker/handlers/etcd"
	"github.com/plexsysio/dLocker/testsuite"
)

func TestLockerSuite(t *testing.T) {
	logger.SetLogLevel("locker/etcd", "Debug")
	l, err := etcd.New("localhost:2379")
	if err != nil {
		t.Fatal("failed starting locker", err)
	}
	defer func() {
		err := l.Close()
		if err != nil {
			t.Fatal("failed closing etcd locker", err)
		}
	}()
	testsuite.RunTests(t, l)
}

func BenchmarkSuite(b *testing.B) {
	logger.SetLogLevel("locker/etcd", "Error")
	l, err := etcd.New("localhost:2379")
	if err != nil {
		b.Fatal("failed starting locker", err)
	}
	defer func() {
		err := l.Close()
		if err != nil {
			b.Fatal("failed closing etcd locker", err)
		}
	}()
	testsuite.RunBenchmark(b, l)
}
