package cors

import (
	"net/http"

	middleware "github.com/SKF/go-enlight-middleware"
)

type Middleware struct{}

// New returns a new cors middleware which act as the default route for all OPTIONS requests and add
// Access-Control-Allow-Origin header to all responses.
func New(opts ...Option) *Middleware {
	m := &Middleware{}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

func (m *Middleware) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, span := middleware.StartSpan(r.Context(), "Middleware/CORS")

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
