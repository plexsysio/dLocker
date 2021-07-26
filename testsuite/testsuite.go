package testsuite

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/plexsysio/dLocker"
)

func getLocker(val interface{}) dLocker.DLocker {
	l, ok := val.(dLocker.DLocker)
	if !ok {
		panic("Pkg does not implement DLocker")
	}
	return l
}

func run(
	t *testing.T,
	factory func() interface{},
	tests ...func(*testing.T, dLocker.DLocker),
) {
	for _, ut := range tests {
		tName := runtime.FuncForPC(reflect.ValueOf(ut).Pointer()).Name()
		tArr := strings.Split(tName, ".")
		fmt.Println("=====\tRUN\t", tArr[len(tArr)-1])
		start := time.Now()
		ut(t, getLocker(factory()))
		fmt.Println("=====\tPASS\t", tArr[len(tArr)-1], "--", time.Since(start).String())
	}
}

func RunTests(t *testing.T, factory func() interface{}) {
	fmt.Println("=====\tDLocker Testsuite")
	run(t, factory,
		TestLock_Close,
		TestLock_TryLockSimple,
		TestLock_TryLockMultiple,
		TestLock_TryLockFailSucceed,
	)
}

func TestLock_Close(t *testing.T, l dLocker.DLocker) {
	if err := l.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestLock_TryLockSimple(t *testing.T, l dLocker.DLocker) {
	defer l.Close()

	unlock, err := l.TryLock(context.Background(), "dummyKey", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Millisecond * 100)
	unlock()
}

func TestLock_TryLockMultiple(t *testing.T, l dLocker.DLocker) {
	defer l.Close()

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
	defer l.Close()

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
