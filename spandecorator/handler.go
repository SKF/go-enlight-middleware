package spandecorator

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/SKF/go-rest-utility/problems"
	"github.com/SKF/go-utility/v2/useridcontext"

	"github.com/SKF/go-enlight-middleware/spandecorator/internal"

	middleware "github.com/SKF/go-enlight-middleware"
)

// Limit tag value to 5000 characters (limit in datadog)
// https://docs.datadoghq.com/tracing/troubleshooting/#data-volume-guidelines
const maxTagValueSize int = 5000

type Middleware struct {
	Tracer   middleware.Tracer
	withBody bool
}

func New(opts ...Option) *Middleware {
	mw := &Middleware{Tracer: middleware.DefaultTracer}

	for _, opt := range opts {
		opt(mw)
	}

	return mw
}

func (m *Middleware) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			span := m.Tracer.SpanFromContext(ctx)

			for k, v := range extractAttributes(r) {
				span.AddStringAttribute(k, v)
			}

			if m.withBody && !emptyBody(r) {
				partialBody, err := extractPartialBody(r, maxTagValueSize)
				if err != nil {
					problems.WriteResponse(ctx, err, w, r)
					return
				}

				span.AddStringAttribute("http.request.body", string(partialBody))
			}

			next.ServeHTTP(w, r)
		})
	}
}

func extractAttributes(r *http.Request) map[string]string {
	attributes := map[string]string{}

	for key, values := range r.Header {
		if shouldIgnore(key) {
			continue
		}

		switch len(values) {
		case 0:
			attributes[fmt.Sprintf("header.%s", key)] = ""
		case 1:
			attributes[fmt.Sprintf("header.%s", key)] = values[0]
		default:
			for i := range values {
				attributes[fmt.Sprintf("header.%s.%d", key, i)] = values[i]
			}
		}
	}

	userID, ok := useridcontext.FromContext(r.Context())
	if ok {
		attributes[internal.UserIDKey] = userID
	}

	return attributes
}

func shouldIgnore(key string) bool {
	if key == "Authorization" {
		return true
	}

	if key == "X-Forwarded-For" {
		return true
	}

	return false
}

func extractPartialBody(r *http.Request, limit int) ([]byte, error) {
	var b []byte
	var err error

	if b, err = io.ReadAll(r.Body); err != nil {
		return nil, fmt.Errorf("unable to read body from request: %w", err)
	}

	if err = r.Body.Close(); err != nil {
		return nil, err
	}

	r.Body = io.NopCloser(bytes.NewReader(b))

	if len(b) > limit {
		b = b[0:limit]
	}

	return b, nil
}

func emptyBody(r *http.Request) bool {
	return r.Body == nil || r.Body == http.NoBody
}
