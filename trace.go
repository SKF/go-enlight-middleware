package middleware

import (
	"context"
	"strings"

	"go.opencensus.io/trace"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace"
	ddtracer "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func StartSpan(ctx context.Context, resourceName string) (context.Context, Span) {
	// Avoid creating a new trace for the middlewares, most requests will have a trace but
	// health endpoints will not. This will avoid creating unnessesary traces for those.
	_, isFound := ddtracer.SpanFromContext(ctx)
	if isFound {
		ddResourceName := parseResourceNameToDDName(resourceName)

		span, ctx := ddtracer.StartSpanFromContext(ctx, "web.middleware", ddtracer.ResourceName(ddResourceName))

		return ctx, datadogSpan{span: span}
	}

	if trace.FromContext(ctx) != nil {
		ctx, span := trace.StartSpan(ctx, resourceName)

		return ctx, openCensusSpan{span: span}
	}

	return ctx, &NilSpan{}
}

func SpanFromContext(ctx context.Context) Span {
	span, isFound := ddtracer.SpanFromContext(ctx)
	if isFound {
		return datadogSpan{span: span}
	}

	if span := trace.FromContext(ctx); span != nil {
		return openCensusSpan{span: span}
	}

	return &NilSpan{}
}

type Span interface {
	End()
	AddStringAttribute(name, value string)
}

type NilSpan struct{}

func (s *NilSpan) End() {}

func (s *NilSpan) AddStringAttribute(name, value string) {}

type openCensusSpan struct {
	span *trace.Span
}

func (s openCensusSpan) End() {
	s.span.End()
}

func (s openCensusSpan) AddStringAttribute(name, value string) {
	s.span.AddAttributes(trace.StringAttribute(name, value))
}

func parseResourceNameToDDName(resourceName string) (ddResourceName string) {
	sParts := strings.Split(resourceName, "/")

	if len(sParts) == 1 {
		ddResourceName = sParts[0]
	} else {
		ddResourceName = strings.Join(sParts[1:], ".")
	}

	return
}

type datadogSpan struct {
	span ddtrace.Span
}

func (s datadogSpan) End() {
	s.span.Finish()
}

func (s datadogSpan) AddStringAttribute(name, value string) {
	s.span.SetTag(name, value)
}
