package trace_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/SKF/go-enlight-middleware/internal/trace"
)

type TracerMock struct {
	mock.Mock
}

type SpanMock struct {
	mock.Mock
}

func (t *TracerMock) StartSpan(ctx context.Context, resourceName string) (context.Context, trace.Span) {
	args := t.Called(ctx, resourceName)
	return args.Get(0).(context.Context), args.Get(1).(trace.Span)
}

func (t *TracerMock) SpanFromContext(ctx context.Context) trace.Span {
	args := t.Called(ctx)
	return args.Get(0).(trace.Span)
}

func (s *SpanMock) End() {
	s.Called()
}

func (s *SpanMock) AddStringAttribute(name, value string) {
	s.Called(name, value)
}

func (s *SpanMock) Empty() bool {
	return s.Called().Bool(0)
}

func (s *SpanMock) Internal() any {
	return s.Called().Get(0)
}

func TestMultiTracerStartSpan_AllEmpty(t *testing.T) {
	ctx := context.Background()

	t1, t2 := new(TracerMock), new(TracerMock)
	t1.On("StartSpan", ctx, "A").Return(context.Background(), new(trace.NilSpan)).Once()
	t2.On("StartSpan", ctx, "A").Return(context.Background(), new(trace.NilSpan)).Once()

	ts := trace.MultiTracer{
		t1, t2,
	}

	_, span := ts.StartSpan(ctx, "A")

	require.IsType(t, new(trace.NilSpan), span)

	t1.AssertExpectations(t)
	t2.AssertExpectations(t)
}

func TestMultiTracerStartSpan_FirstNotEmpty(t *testing.T) {
	ctx := context.Background()

	s1 := new(SpanMock)
	s1.On("Empty").Return(false).Once()

	t1, t2 := new(TracerMock), new(TracerMock)
	t1.On("StartSpan", ctx, "A").Return(context.Background(), s1).Once()

	ts := trace.MultiTracer{
		t1, t2,
	}

	_, span := ts.StartSpan(ctx, "A")

	require.Equal(t, s1, span)

	s1.AssertExpectations(t)
	t1.AssertExpectations(t)
	t2.AssertExpectations(t)
}

func TestMultiTracerStartSpan_FirstEmpty(t *testing.T) {
	ctx := context.Background()

	s1 := new(SpanMock)
	s1.On("Empty").Return(false).Once()

	t1, t2 := new(TracerMock), new(TracerMock)
	t1.On("StartSpan", ctx, "A").Return(context.Background(), new(trace.NilSpan)).Once()
	t2.On("StartSpan", ctx, "A").Return(context.Background(), s1).Once()

	ts := trace.MultiTracer{
		t1, t2,
	}

	_, span := ts.StartSpan(ctx, "A")

	require.Equal(t, s1, span)

	s1.AssertExpectations(t)
	t1.AssertExpectations(t)
	t2.AssertExpectations(t)
}
