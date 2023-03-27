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

type Middleware struct {
	Tracer   middleware.Tracer
	withBody bool
}

func New() *Middleware {
	return &Middleware{Tracer: middleware.DefaultTracer}
}

func (m *Middleware) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			span := m.Tracer.SpanFromContext(ctx)

			for k, v := range extractAttributes(r) {
				span.AddStringAttribute(k, v)
			}

			if m.withBody {
				if err := decorateWithBody(r, span); err != nil {
					problems.WriteResponse(ctx, err, w, r)
				}
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

func decorateWithBody(r *http.Request, span middleware.Span) (err error) {
	if r.Body == nil || r.Body == http.NoBody {
		return nil
	}

	var logBody io.ReadCloser

	logBody, r.Body, err = drainBody(r.Body)
	if err != nil {
		return fmt.Errorf("unable to extract body from request: %w", err)
	}

	//Limit tag value to 5000 characters(limit in datadog)
	b, err := io.ReadAll(io.LimitReader(logBody, 5000)) //nolint: gomnd
	if err != nil {
		return fmt.Errorf("unable to read body from request: %w", err)
	}

	span.AddStringAttribute("http.request.body", string(b))

	return nil
}

func drainBody(b io.ReadCloser) (r1, r2 io.ReadCloser, err error) {
	var buf bytes.Buffer
	if _, err = buf.ReadFrom(b); err != nil {
		return nil, b, err
	}

	if err = b.Close(); err != nil {
		return nil, b, err
	}

	return io.NopCloser(&buf), io.NopCloser(bytes.NewReader(buf.Bytes())), nil
}
