package trace

import "context"

// MultiTracer uses the first tracer that responds with a non-empty span
type MultiTracer []Tracer

func (ts MultiTracer) StartSpan(ctx context.Context, resourceName string) (context.Context, Span) {
	for _, t := range ts {
		if newCtx, span := t.StartSpan(ctx, resourceName); !span.Empty() {
			return newCtx, span
		}
	}

	return ctx, new(NilSpan)
}

func (ts MultiTracer) SpanFromContext(ctx context.Context) Span {
	for _, t := range ts {
		if span := t.SpanFromContext(ctx); !span.Empty() {
			return span
		}
	}

	return new(NilSpan)
}
