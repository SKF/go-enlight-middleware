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
	if trace.FromContext(ctx) == nil {
		return ctx, nil
	}

	return trace.StartSpan(ctx, resourceName)
}
