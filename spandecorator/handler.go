package spandecorator

import (
	"fmt"
	"net/http"

	"github.com/SKF/go-utility/v2/log"
	"github.com/SKF/go-utility/v2/useridcontext"

	middleware "github.com/SKF/go-enlight-middleware"
)

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
			for key, values := range r.Header {
				if key == "Authorization" {
					continue
				}

				switch len(values) {
				case 0:
					span.AddStringAttribute(fmt.Sprintf("header.%s", key), "")
				case 1:
					span.AddStringAttribute(fmt.Sprintf("header.%s", key), values[0])
				default:
					for i := range values {
						span.AddStringAttribute(fmt.Sprintf("header.%s.%d", key, i), values[i])
					}
				}
			}

			userID, ok := useridcontext.FromContext(r.Context())
			if !ok {
				log.WithTracing(r.Context()).
					Warning("Failed to get userID from context")
			}

			span.AddStringAttribute("callerId", userID)

			next.ServeHTTP(w, r)
		})
	}
}
