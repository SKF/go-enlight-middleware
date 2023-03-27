package trace_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	oc_trace "go.opencensus.io/trace"
	dd_tracer "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

	"github.com/SKF/go-enlight-middleware/internal/trace"
)

func TestDatadog_RootSpanFromContext(t *testing.T) {
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
	_, ctx := dd_tracer.StartSpanFromContext(context.Background(), "web", dd_tracer.ResourceName("GET /"))

	_, span := new(trace.DatadogTracer).StartSpan(ctx, "A")

	require.IsType(t, new(oc_trace.Span), span.Internal())
	internal := span.Internal().(*oc_trace.Span)

	require.Contains(t, internal.String(), `"Middleware/A"`)
}
