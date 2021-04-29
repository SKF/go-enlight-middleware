package cors

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type preflightInfo struct {
	allowedMethods []string
	allowedHeaders []string
}

type Middleware struct {
	paths map[string]preflightInfo
}

func New() *Middleware {
	return &Middleware{
		paths: map[string]preflightInfo{},
	}
}

func (m *Middleware) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
			if r.Method == http.MethodOptions {
				path, err := mux.CurrentRoute(r).GetPathTemplate()
				if err != nil {
					return
				}

				if routeInfo, found := m.paths[path]; found {
					allowedMethods := strings.Join(routeInfo.allowedMethods, ", ")
					allowedHeaders := strings.Join(routeInfo.allowedHeaders, ", ")

					w.Header().Set("Access-Control-Allow-Methods", allowedMethods)
					w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
				}
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (m *Middleware) AddAllowedHeaders(route *mux.Route, headers ...string) *Middleware {
	if path, err := route.GetPathTemplate(); err == nil {
		preflight := m.paths[path]
		preflight.allowedHeaders = append(preflight.allowedHeaders, headers...)

		m.paths[path] = preflight
	}

	return m
}

func (m *Middleware) AddAllowedMethods(route *mux.Route, methods ...string) *Middleware {
	if path, err := route.GetPathTemplate(); err == nil {
		preflight := m.paths[path]
		preflight.allowedMethods = append(preflight.allowedMethods, methods...)

		m.paths[path] = preflight
	}

	return m
}
