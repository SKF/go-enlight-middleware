package spandecorator

import (
	"fmt"
	"net/http"

	"github.com/SKF/go-utility/v2/useridcontext"

	"github.com/SKF/go-enlight-middleware/spandecorator/internal"

	middleware "github.com/SKF/go-enlight-middleware"
)

type Middleware struct {
	Tracer middleware.Tracer
}

func New() *Middleware {
	return &Middleware{Tracer: middleware.DefaultTracer}
}

func (m *Middleware) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			span := m.Tracer.SpanFromContext(r.Context())

			for k, v := range extractAttributes(r) {
				span.AddStringAttribute(k, v)
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
