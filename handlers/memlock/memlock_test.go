package memlock_test

import (
	"testing"

	"github.com/plexsysio/dLocker/handlers/memlock"
	"github.com/plexsysio/dLocker/testsuite"
)

func TestLockerSuite(t *testing.T) {
	testsuite.RunTests(t, memlock.NewLocker())
}

func BenchmarkSuite(b *testing.B) {
	testsuite.RunBenchmark(b, memlock.NewLocker())
}
