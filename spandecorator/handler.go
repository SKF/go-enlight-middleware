package spandecorator

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/SKF/go-utility/v2/log"
	"github.com/SKF/go-utility/v2/useridcontext"

	middleware "github.com/SKF/go-enlight-middleware"
)

const UserIDKey = "userId"

type Middleware struct {
	Tracer middleware.Tracer
}

func New() *Middleware {
	return &Middleware{Tracer: &middleware.OpenCensusTracer{}}
}

func (m *Middleware) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			span := m.Tracer.SpanFromContext(r.Context())
			attrs := getAttributes(r)
			for k, v := range attrs {
				span.AddStringAttribute(k, v)
			}

			next.ServeHTTP(w, r)
		})
	}
}

func getAttributes(r *http.Request) map[string]string {
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
	if !ok {
		log.WithTracing(r.Context()).
			Warning("Failed to get userID from context")
	}

	attributes[UserIDKey] = userID

	return attributes
}

func shouldIgnore(key string) bool {
	if key == "Authorization" {
		return true
	}

	if strings.ToLower(key) == "x-forwarded-for" {
		return true
	}

	return false
}
