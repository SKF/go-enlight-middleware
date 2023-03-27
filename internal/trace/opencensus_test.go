package trace_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	oc_trace "go.opencensus.io/trace"

	"github.com/SKF/go-enlight-middleware/internal/trace"
)

func TestOpenCensus_RootSpanFromContext(t *testing.T) {
	ctx, root := oc_trace.StartSpan(context.Background(), "root")

	span := new(trace.OpenCensusTracer).SpanFromContext(ctx)

	require.Same(t, root, span.Internal())
}

func TestOpenCensus_SpanFromNilContext(t *testing.T) {
	span := new(trace.OpenCensusTracer).SpanFromContext(context.Background())

	require.IsType(t, new(trace.NilSpan), span)
}

func TestOpenCensus_StartSpanNilSpan(t *testing.T) {
	_, span := new(trace.OpenCensusTracer).StartSpan(context.Background(), "A")

	require.IsType(t, new(trace.NilSpan), span)
}

func TestOpenCensus_StartSpan(t *testing.T) {
	ctx, _ := oc_trace.StartSpan(context.Background(), "root", oc_trace.WithSampler(oc_trace.AlwaysSample()))

	_, span := new(trace.OpenCensusTracer).StartSpan(ctx, "A")

	require.IsType(t, new(oc_trace.Span), span.Internal())
	internal := span.Internal().(*oc_trace.Span)

	require.Contains(t, internal.String(), `"Middleware/A"`)
}
