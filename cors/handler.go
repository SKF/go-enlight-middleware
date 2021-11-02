package cors

import (
	"net/http"

	middleware "github.com/SKF/go-enlight-middleware"
)

type Middleware struct {
	Tracer middleware.Tracer
}

// New returns a new cors middleware which adds cors headers to the responses.
func New(opts ...Option) *Middleware {
	m := &Middleware{
		Tracer: new(middleware.OpenCensusTracer),
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

func (m *Middleware) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, span := m.Tracer.StartSpan(r.Context(), "Middleware/CORS")

			w.Header().Set("Access-Control-Allow-Origin", "*")
			if r.Method == http.MethodOptions {
				w.Header().Set("Access-Control-Allow-Methods", "*")
				w.Header().Set("Access-Control-Allow-Headers", "*")

				span.End()
				return
			}

			span.End()
			next.ServeHTTP(w, r)
		})
	}
}
