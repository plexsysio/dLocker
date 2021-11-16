package testsuite

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/plexsysio/dLocker"
)

func RunBenchmark(b *testing.B, l dLocker.DLocker) {
	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		unlock, err := l.TryLock(context.TODO(), fmt.Sprintf("test%d", n), time.Second)
		if err != nil {
			b.Fatal("failed to lock", n)
		}
		unlock()
	}
}
