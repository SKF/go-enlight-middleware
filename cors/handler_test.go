package cors

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/SKF/go-enlight-middleware/cors/preflight"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

type testRoute struct {
	path    string
	methods []string
}

type testCase struct {
	name            string
	testRoutes      []testRoute
	requestPath     string
	requestOrigin   string
	allowedHeaders  []string
	expectedHeaders map[string][]string
	actualHeaders   map[string][]string
}

func setupAndDoRequest(t *testing.T, tc *testCase) {
	endpoint := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		assert.Fail(t, "Reached endpoint handler other than the OPTIONS handler")
	})

	request := httptest.NewRequest(http.MethodOptions, tc.requestPath, nil)
	if tc.requestOrigin != "" {
		request.Header.Add("Origin", tc.requestOrigin)
	}

	w := httptest.NewRecorder()

	router := mux.NewRouter()
	router.Use(Middleware())

	for _, r := range tc.testRoutes {
		r := r
		router.NewRoute().
			Methods(r.methods...).
			Path(r.path).
			HandlerFunc(endpoint)
	}

	preflight.AddHandler(router, tc.allowedHeaders...)

	router.ServeHTTP(w, request)

	response := w.Result()
	defer response.Body.Close()
	tc.actualHeaders = normalizeHeaders(response.Header)
}

func normalizeHeaders(headers map[string][]string) (normalized map[string][]string) {
	normalized = make(map[string][]string)

	for header, values := range headers {
		headerValues := make([]string, 0)
		for _, value := range values {
			headerValues = append(headerValues, strings.Split(value, ", ")...)
		}

		normalized[header] = headerValues
	}

	return
}

func TestCORSPreflightHeaders(t *testing.T) {
	testCases := []testCase{
		{
			name: "Allow-Origin should be set if included in the request",
			testRoutes: []testRoute{
				{
					path:    "/get",
					methods: []string{"GET"},
				},
			},
			requestPath:   "/get",
			requestOrigin: "testOrigin",

			expectedHeaders: map[string][]string{
				"Access-Control-Allow-Methods": {"GET"},
				"Access-Control-Allow-Origin":  {"testOrigin"},
			},
		},
		{
			name: "Only return methods for requested endpoint",
			testRoutes: []testRoute{
				{
					path:    "/get",
					methods: []string{"GET"},
				},
				{
					path:    "/put",
					methods: []string{"PUT"},
				},
			},
			requestPath:    "/get",
			allowedHeaders: []string{"Test-Header-1", "Test-Header-2"},

			expectedHeaders: map[string][]string{
				"Access-Control-Allow-Headers": {"Test-Header-1", "Test-Header-2"},
				"Access-Control-Allow-Methods": {"GET"},
				"Access-Control-Allow-Origin":  {"*"},
			},
		},
		{
			name: "Return all methods for the given path, even if different route instances",
			testRoutes: []testRoute{
				{
					path:    "/",
					methods: []string{"GET"},
				},
				{
					path:    "/",
					methods: []string{"PUT", "PATCH"},
				},
			},
			requestPath:    "/",
			allowedHeaders: []string{"Test-Header-Get", "Test-Header-PutPatch"},

			expectedHeaders: map[string][]string{
				"Access-Control-Allow-Headers": {"Test-Header-Get", "Test-Header-PutPatch"},
				"Access-Control-Allow-Methods": {"GET", "PUT", "PATCH"},
				"Access-Control-Allow-Origin":  {"*"},
			},
		},
		{
			name: "Paths with regexp should be handled correctly",
			testRoutes: []testRoute{
				{
					path:    "/nodes/{node:[a-zA-Z0-9-]+}",
					methods: []string{"GET", "PUT", "PATCH", "DELETE"},
				},
			},
			requestPath:    "/nodes/19b27d52-2e71-416f-a1c3-8e4e9d43e691",
			allowedHeaders: []string{"Test-Header-Get", "Test-Header-PutPatch"},

			expectedHeaders: map[string][]string{
				"Access-Control-Allow-Headers": {"Test-Header-Get", "Test-Header-PutPatch"},
				"Access-Control-Allow-Methods": {"GET", "PUT", "PATCH", "DELETE"},
				"Access-Control-Allow-Origin":  {"*"},
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		setupAndDoRequest(t, &tc)

		assert.Equal(t, tc.expectedHeaders, tc.actualHeaders, "Test %s returned unexpected headers", tc.name)
	}
}
