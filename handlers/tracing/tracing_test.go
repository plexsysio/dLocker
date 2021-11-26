package tracing_test

import (
	"strings"
	"testing"

	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/plexsysio/dLocker/handlers/memlock"
	"github.com/plexsysio/dLocker/handlers/tracing"
	"github.com/plexsysio/dLocker/testsuite"
)

func TestLockerSuite(t *testing.T) {
	tracer := mocktracer.New()

	testsuite.RunTests(t, tracing.New(memlock.NewLocker(), tracer))

	// Following values are based on the testsuite operations
	if len(tracer.FinishedSpans()) != 7 {
		t.Fatal("incorrect no of spans", len(tracer.FinishedSpans()))
	}
	ops := 0
	for _, v := range tracer.FinishedSpans() {
		if strings.HasPrefix(v.OperationName, "TryLock") {
			ops++
		}
	}
	if ops != 7 {
		t.Fatal("ops count incorrect", ops)
	}
}

func BenchmarkSuite(b *testing.B) {
	tracer := mocktracer.New()

	testsuite.RunBenchmark(b, tracing.New(memlock.NewLocker(), tracer))
}
