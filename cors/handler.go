package cors

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type Methods []string

func handler(allowedMethods, allowedHeaders []string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		allowedMethods := strings.Join(allowedMethods, ", ")
		allowedHeaders := strings.Join(allowedHeaders, ", ")

		w.Header().Set("Access-Control-Allow-Methods", allowedMethods)
		w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
	}
}

func AddCORSHandler(router *mux.Router, allowedHeaders ...string) {
	pathMethods := make(map[string]Methods)
	routeWalker := func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		path, err := route.GetPathTemplate()
		if err != nil {
			return nil
		}

		routeMethods, err := route.GetMethods()
		if err != nil {
			return nil
		}

		methods, found := pathMethods[path]
		if !found {
			methods = Methods{}
		}

		filter := func(method string) bool { return method != http.MethodOptions }

		pathMethods[path] = addMissingEntries(methods, routeMethods, filter)

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
