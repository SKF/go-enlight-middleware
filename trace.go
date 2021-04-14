package middleware

import (
	"context"

	"go.opencensus.io/trace"
)

type Tracer interface {
	StartSpan(ctx context.Context, resourceName string) (context.Context, Span)
}

type Span interface {
	End()
	AddStringAttribute(name, value string)
}

type NilSpan struct{}

func (s *NilSpan) End() {}

func (s *NilSpan) AddStringAttribute(name, value string) {}

type OpenCensusTracer struct{}

type enlightSpan struct {
	span *trace.Span
}

func (s enlightSpan) End() {
	s.span.End()
}

func (s enlightSpan) AddStringAttribute(name, value string) {
	s.span.AddAttributes(trace.StringAttribute(name, value))
}

func (t *OpenCensusTracer) StartSpan(ctx context.Context, resourceName string) (context.Context, Span) {
	// Avoid creating a new trace for the middlewares, most requests will have a trace but
	// health endpoints will not. This will avoid creating unnessesary traces for those.
	if trace.FromContext(ctx) == nil {
		return ctx, &NilSpan{}
	}

	ctx, span := trace.StartSpan(ctx, resourceName)

	return ctx, enlightSpan{span: span}
}
