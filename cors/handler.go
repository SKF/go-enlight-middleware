package cors

import (
	"fmt"
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
		preflight.allowedHeaders = addMissingEntries(preflight.allowedHeaders, headers, nil)

		m.paths[path] = preflight
		fmt.Printf("preflight: %+v\n", preflight)
		fmt.Printf("m.paths[path]: %+v\n", m.paths[path])
	}

	return m
}

func (m *Middleware) AddAllowedMethods(route *mux.Route, methods ...string) *Middleware {
	if path, err := route.GetPathTemplate(); err == nil {
		preflight := m.paths[path]
		filter := func(method string) bool { return method != http.MethodOptions }
		preflight.allowedMethods = addMissingEntries(preflight.allowedMethods, methods, filter)

		m.paths[path] = preflight
	}

	return m
}

func addMissingEntries(list, newEntries []string, filter func(string) bool) []string {
	for _, e := range newEntries {
		if !contains(list, e) && (filter == nil || filter(e)) {
			list = append(list, e)
		}
	}

	return list
}

func contains(list []string, entry string) bool {
	for _, e := range list {
		if e == entry {
			return true
		}
	}

	return false
}
