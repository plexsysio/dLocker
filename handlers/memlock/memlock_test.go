package memlock

import (
	"github.com/aloknerurkar/dLocker/testsuite"
	"testing"
)

func TestLockerSuite(t *testing.T) {
	testsuite.RunTests(t, func() interface{} {
		return NewLocker()
	})
}
