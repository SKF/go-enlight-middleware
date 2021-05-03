package cors

import "net/http"

func Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := "*"
			if explicitOrigin := r.Header.Get("Origin"); explicitOrigin != "" {
				origin = explicitOrigin
			}
			w.Header().Set("Access-Control-Allow-Origin", origin)

			next.ServeHTTP(w, r)
		})
	}
}
