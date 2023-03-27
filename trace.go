package middleware

import "github.com/SKF/go-enlight-middleware/internal/trace"

type (
	Tracer = trace.Tracer
	Span   = trace.Span

	// Legacy, reexported for backwards compatibility
	OpenCensusTracer = trace.OpenCensusTracer
	NilSpan          = trace.NilSpan
)

var DefaultTracer Tracer = trace.MultiTracer{
	new(trace.DatadogTracer),
	new(trace.OpenCensusTracer),
}
