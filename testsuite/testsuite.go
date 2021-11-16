package testsuite

import (
	"context"
	"testing"
	"time"

	"github.com/plexsysio/dLocker"
)

func RunTests(t *testing.T, impl dLocker.DLocker) {
	t.Run("Simple", func(st *testing.T) {
		TestLock_TryLockSimple(st, impl)
	})
	t.Run("Multiple", func(st *testing.T) {
		TestLock_TryLockMultiple(st, impl)
	})
	t.Run("FailSucceed", func(st *testing.T) {
		TestLock_TryLockFailSucceed(st, impl)
	})
}

func TestLock_TryLockSimple(t *testing.T, l dLocker.DLocker) {
	unlock, err := l.TryLock(context.Background(), "dummyKey", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Millisecond * 100)
	unlock()
}

func TestLock_TryLockMultiple(t *testing.T, l dLocker.DLocker) {
	for i := 0; i < 3; i++ {
		t.Run("TryLock", func(t *testing.T) {
			unlock, err := l.TryLock(context.Background(), "dummyKey", time.Second*3)
			if err != nil {
				t.Fatal(err)
			}
			time.Sleep(time.Millisecond * 100)
			unlock()
		})
	}
}

func TestLock_TryLockFailSucceed(t *testing.T, l dLocker.DLocker) {
	c := context.Background()

	unlock, err := l.TryLock(c, "dummyKey", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := l.TryLock(c, "dummyKey", 15*time.Second); err == nil {
		t.Fatal("Able to obtain lock when its not unlocked")
	}
	unlock()
	time.Sleep(time.Millisecond * 100)

	if unlock, err := l.TryLock(c, "dummyKey", time.Second); err != nil {
		t.Fatal(err)
	} else {
		unlock()
	}
}
