package recovery

import (
	"fmt"
	"net/http"

	"github.com/SKF/go-rest-utility/problems"
)

type Middleware struct{}

func New() *Middleware {
	return &Middleware{}
}

func (m *Middleware) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if panic := recover(); panic != nil {
					err, ok := panic.(error)
					if !ok {
						err = fmt.Errorf("panic: %v", panic)
					}

					problems.WriteResponse(r.Context(), err, w, r)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
