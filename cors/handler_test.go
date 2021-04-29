package cors

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

type testRoute struct {
	path    string
	methods []string
	headers []string
}

type testCase struct {
	testRoutes      []testRoute
	requestPath     string
	expectedHeaders []string
	expectedMethods []string
	actualHeaders   []string
	actualMethods   []string
}

func setupAndDoRequest(t *testing.T, tc *testCase, middleware *Middleware) {
	endpoint := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		assert.Fail(t, "Reached endpoint handler")
	})

	request := httptest.NewRequest(http.MethodOptions, tc.requestPath, nil)
	w := httptest.NewRecorder()

	router := mux.NewRouter()
	router.Use(middleware.Middleware())

	for _, r := range tc.testRoutes {
		r := r
		route := router.NewRoute().
			Methods(r.methods...).
			Path(r.path).
			HandlerFunc(endpoint)
		middleware.AddAllowedMethods(route, r.methods...)
		middleware.AddAllowedHeaders(route, r.headers...)
	}

	router.ServeHTTP(w, request)

	response := w.Result()
	defer response.Body.Close()
	tc.actualHeaders = strings.Split(response.Header.Get("Access-Control-Allow-Headers"), ", ")
	tc.actualMethods = strings.Split(response.Header.Get("Access-Control-Allow-Methods"), ", ")
}

func TestCORSPreflightHeaders(t *testing.T) {
	testCases := []testCase{
		{
			testRoutes: []testRoute{
				{
					path:    "/get",
					methods: []string{"GET", "OPTIONS"},
					headers: []string{"Test-Header-Get"},
				},
				{
					path:    "/put",
					methods: []string{"PUT", "OPTIONS"},
					headers: []string{"Test-Header-Put"},
				},
			},
			requestPath:     "/get",
			expectedMethods: []string{"GET"},
			expectedHeaders: []string{"Test-Header-Get"},
		},
		{
			testRoutes: []testRoute{
				{
					path:    "/",
					methods: []string{"GET", "OPTIONS"},
					headers: []string{"Test-Header-Get"},
				},
				{
					path:    "/",
					methods: []string{"PUT", "PATCH", "OPTIONS"},
					headers: []string{"Test-Header-PutPatch"},
				},
			},
			requestPath:     "/",
			expectedMethods: []string{"GET", "PUT", "PATCH"},
			expectedHeaders: []string{"Test-Header-Get", "Test-Header-PutPatch"},
		},
	}
	for _, tc := range testCases {
		tc := tc
		setupAndDoRequest(t, &tc, New())

		assert.Equal(t, tc.expectedHeaders, tc.actualHeaders)
		assert.Equal(t, tc.expectedMethods, tc.actualMethods)
	}
}
