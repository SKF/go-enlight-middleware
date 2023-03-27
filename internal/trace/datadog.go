package trace

import (
	"context"

	dd_trace "gopkg.in/DataDog/dd-trace-go.v1/ddtrace"
	dd_tracer "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type DatadogTracer struct{}

type datadogSpan struct {
	span dd_trace.Span
}

func (s datadogSpan) End() {
	s.span.Finish()
}

func (s datadogSpan) AddStringAttribute(name, value string) {
	s.span.SetTag(name, value)
}

func (s datadogSpan) Empty() bool {
	return false
}

func (s datadogSpan) Internal() any {
	return s.span
}

func (t *DatadogTracer) StartSpan(ctx context.Context, resourceName string) (context.Context, Span) {
	if emptySpan := t.SpanFromContext(ctx); emptySpan.Empty() {
		return ctx, emptySpan
	}

	span, ctx := dd_tracer.StartSpanFromContext(ctx, "web.middleware", dd_tracer.ResourceName(resourceName))

	return ctx, datadogSpan{span}
}

func (t *DatadogTracer) SpanFromContext(ctx context.Context) Span {
	span, found := dd_tracer.SpanFromContext(ctx)
	if !found {
		return new(NilSpan)
	}

	return datadogSpan{span}
}
