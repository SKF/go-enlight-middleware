package middleware

import (
	"context"
	"strings"

	"go.opencensus.io/trace"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace"
	ddtracer "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type Tracer interface {
	StartSpan(ctx context.Context, resourceName string) (context.Context, Span)
	SpanFromContext(ctx context.Context) Span
}

type Span interface {
	End()
	AddStringAttribute(name, value string)
}

type NilSpan struct{}

func (s *NilSpan) End() {}

func (s *NilSpan) AddStringAttribute(name, value string) {}

type OpenCensusTracer struct{}

type openCensusSpan struct {
	span *trace.Span
}

func (s openCensusSpan) End() {
	s.span.End()
}

func (s openCensusSpan) AddStringAttribute(name, value string) {
	s.span.AddAttributes(trace.StringAttribute(name, value))
}

func (t *OpenCensusTracer) StartSpan(ctx context.Context, resourceName string) (context.Context, Span) { // Avoid creating a new trace for the middlewares, most requests will have a trace but
	// health endpoints will not. This will avoid creating unnessesary traces for those.
	if trace.FromContext(ctx) == nil {
		return ctx, &NilSpan{}
	}

	ctx, span := trace.StartSpan(ctx, resourceName)

	return ctx, openCensusSpan{span: span}
}

func (t *OpenCensusTracer) SpanFromContext(ctx context.Context) Span {
	span := trace.FromContext(ctx)
	if span == nil {
		return &NilSpan{}
	}

	return openCensusSpan{span: span}
}

type DataDogTracer struct{}

func (t *DataDogTracer) SpanFromContext(ctx context.Context) Span {
	span, _ := ddtracer.SpanFromContext(ctx)
	if span == nil {
		return &NilSpan{}
	}

	return datadogSpan{span: span}
}

func (t *DataDogTracer) StartSpan(ctx context.Context, resourceName string) (context.Context, Span) {
	// health endpoints will not. This will avoid creating unnessesary traces for those.
	_, isFound := ddtracer.SpanFromContext(ctx)
	if !isFound {
		return ctx, &NilSpan{}
	}

	ddOperationName, ddResourceName := parseResourceNameToDDNames(resourceName)

	span, ctx := ddtracer.StartSpanFromContext(ctx, ddOperationName, ddtracer.ResourceName(ddResourceName))

	return ctx, datadogSpan{span: span}
}

func parseResourceNameToDDNames(resourceName string) (ddOperationName, ddResourceName string) {
	sParts := strings.Split(resourceName, "/")

	if len(sParts) == 1 {
		ddOperationName = strings.ToLower(sParts[0])
		ddResourceName = sParts[0]
	} else {
		ddOperationName = strings.ToLower(strings.Join(sParts[:len(sParts)-1], "."))
		ddResourceName = sParts[len(sParts)-1]
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
