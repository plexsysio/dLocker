package tracing

import (
	"context"
	"time"

	"github.com/opentracing/opentracing-go"
	tlog "github.com/opentracing/opentracing-go/log"
	"github.com/plexsysio/dLocker"
)

type tracingLocker struct {
	dLocker.DLocker

	tracer opentracing.Tracer
}

func New(locker dLocker.DLocker, tracer opentracing.Tracer) *tracingLocker {
	return &tracingLocker{
		DLocker: locker,
		tracer:  tracer,
	}
}

func (t *tracingLocker) startSpan(ctx context.Context, operation string, opts ...opentracing.StartSpanOption) opentracing.Span {
	if pCtx := opentracing.SpanFromContext(ctx); pCtx != nil {
		opts = append(opts, opentracing.ChildOf(pCtx.Context()))
	}
	return t.tracer.StartSpan(operation, opts...)
}

func (t *tracingLocker) TryLock(
	ctx context.Context,
	key string,
	timeout time.Duration,
) (func(), error) {

	span := t.startSpan(ctx, "TryLock_"+key)

	unlock, err := t.DLocker.TryLock(ctx, key, timeout)
	if err != nil {
		span.LogFields(tlog.String("error", err.Error()))
		span.Finish()
		return nil, err
	}

	return func() {
		span.Finish()
		unlock()
	}, nil
}
