package trace_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	dd_tracer_mock "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/mocktracer"
	dd_tracer "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

	"github.com/SKF/go-enlight-middleware/internal/trace"
)

func TestDatadog_RootSpanFromContext(t *testing.T) {
	mt := dd_tracer_mock.Start()
	defer mt.Stop()

	root, ctx := dd_tracer.StartSpanFromContext(context.Background(), "web", dd_tracer.ResourceName("GET /"))

	span := new(trace.DatadogTracer).SpanFromContext(ctx)

	require.Same(t, root, span.Internal())
}

func TestDatadog_SpanFromNilContext(t *testing.T) {
	span := new(trace.DatadogTracer).SpanFromContext(context.Background())

	require.IsType(t, new(trace.NilSpan), span)
}

func TestDatadog_StartSpanNilSpan(t *testing.T) {
	_, span := new(trace.DatadogTracer).StartSpan(context.Background(), "A")

	require.IsType(t, new(trace.NilSpan), span)
}

func TestDatadog_StartSpan(t *testing.T) {
	mt := dd_tracer_mock.Start()
	defer mt.Stop()

	_, ctx := dd_tracer.StartSpanFromContext(context.Background(), "web", dd_tracer.ResourceName("GET /"))

	_, span := new(trace.DatadogTracer).StartSpan(ctx, "A")

	require.Implements(t, new(dd_tracer_mock.Span), span.Internal())
	mock := span.Internal().(dd_tracer_mock.Span)

	require.Equal(t, "web.middleware", mock.OperationName())
	require.Equal(t, "A", mock.Tag("resource.name"))
}
