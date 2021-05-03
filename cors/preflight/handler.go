package preflight

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type Methods []string

func handler(allowedMethods, allowedHeaders []string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if allowedMethods := strings.Join(allowedMethods, ", "); allowedMethods != "" {
			w.Header().Set("Access-Control-Allow-Methods", allowedMethods)
		}

		if allowedHeaders := strings.Join(allowedHeaders, ", "); allowedHeaders != "" {
			w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
		}
	}
}

func AddHandler(router *mux.Router, allowedHeaders ...string) {
	pathMethods := make(map[string]Methods)
	routeWalker := func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		routeMethods, err := route.GetMethods()
		if err != nil {
			return nil
		}

		path, err := route.GetPathTemplate()
		if err != nil {
			return nil
		}

		methodsForPath, found := pathMethods[path]
		if !found {
			methodsForPath = Methods{}
		}

		filter := func(method string) bool { return method != http.MethodOptions }

		pathMethods[path] = addMissingEntries(methodsForPath, routeMethods, filter)

		return nil
	}

	router.Walk(routeWalker) // nolint:errcheck

	for path, allowedMethods := range pathMethods {
		router.NewRoute().
			Path(path).
			Methods(http.MethodOptions).
			HandlerFunc(handler(allowedMethods, allowedHeaders))
	}
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
