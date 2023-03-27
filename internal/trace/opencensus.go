package trace

import (
	"context"

	oc_trace "go.opencensus.io/trace"
)

type OpenCensusTracer struct{}

type openCensusSpan struct {
	span *oc_trace.Span
}

func (s openCensusSpan) End() {
	s.span.End()
}

func (s openCensusSpan) AddStringAttribute(name, value string) {
	s.span.AddAttributes(oc_trace.StringAttribute(name, value))
}

func (s openCensusSpan) Empty() bool {
	return false
}

func (s openCensusSpan) Internal() any {
	return s.span
}

func (t *OpenCensusTracer) StartSpan(ctx context.Context, resourceName string) (context.Context, Span) {
	// Avoid creating a new trace for the middlewares, most requests will have a trace but
	// health endpoints will not. This will avoid creating unnessesary traces for those.
	if oc_trace.FromContext(ctx) == nil {
		return ctx, &NilSpan{}
	}

	ctx, span := oc_trace.StartSpan(ctx, "Middleware/"+resourceName)

	return ctx, openCensusSpan{span: span}
}

func (t *OpenCensusTracer) SpanFromContext(ctx context.Context) Span {
	span := oc_trace.FromContext(ctx)
	if span == nil {
		return &NilSpan{}
	}

	return openCensusSpan{span: span}
}
