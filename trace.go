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
}

type OpenCensusTracer struct{}

func (t *OpenCensusTracer) StartSpan(ctx context.Context, resourceName string) (context.Context, Span) {
	// Avoid creating a new trace for the middlewares, most requests will have a trace but
	// health endpoints will not. This will avoid creating unnessesary traces for those.
	if trace.FromContext(ctx) == nil {
		return ctx, nil
	}

	return trace.StartSpan(ctx, resourceName)
}
