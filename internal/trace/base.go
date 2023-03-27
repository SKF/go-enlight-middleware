package trace

import "context"

type Tracer interface {
	StartSpan(ctx context.Context, resourceName string) (context.Context, Span)
	SpanFromContext(ctx context.Context) Span
}

type Span interface {
	End()
	AddStringAttribute(name, value string)

	Empty() bool
	Internal() any
}

type NilSpan struct{}

func (s *NilSpan) End() {}

func (s *NilSpan) AddStringAttribute(name, value string) {}

func (s *NilSpan) Empty() bool {
	return true
}

func (s *NilSpan) Internal() any {
	return s
}
