package memlock

import (
	"github.com/plexsysio/dLocker/testsuite"
	"testing"
)

func TestLockerSuite(t *testing.T) {
	testsuite.RunTests(t, func() interface{} {
		return NewLocker()
	})
}
