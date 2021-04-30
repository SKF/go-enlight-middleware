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
}

type testCase struct {
	testRoutes      []testRoute
	requestPath     string
	allowedHeaders  []string
	expectedMethods []string
	actualHeaders   []string
	actualMethods   []string
}

func setupAndDoRequest(t *testing.T, tc *testCase) {
	endpoint := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		assert.Fail(t, "Reached endpoint handler other than the OPTIONS handler")
	})

	request := httptest.NewRequest(http.MethodOptions, tc.requestPath, nil)
	w := httptest.NewRecorder()

	router := mux.NewRouter()

	for _, r := range tc.testRoutes {
		r := r
		router.NewRoute().
			Methods(r.methods...).
			Path(r.path).
			HandlerFunc(endpoint)
	}

	AddCORSHandler(router, tc.allowedHeaders...)

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
					methods: []string{"GET"},
				},
				{
					path:    "/put",
					methods: []string{"PUT"},
				},
			},
			requestPath:     "/get",
			allowedHeaders:  []string{"Test-Header-1", "Test-Header-2"},
			expectedMethods: []string{"GET"},
		},
		{
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
			requestPath:     "/",
			allowedHeaders:  []string{"Test-Header-Get", "Test-Header-PutPatch"},
			expectedMethods: []string{"GET", "PUT", "PATCH"},
		},
		{
			testRoutes: []testRoute{
				{
					path:    "/nodes/{node:[a-zA-Z0-9-]+}",
					methods: []string{"GET", "PUT", "PATCH", "DELETE"},
				},
			},
			requestPath:     "/nodes/19b27d52-2e71-416f-a1c3-8e4e9d43e691",
			allowedHeaders:  []string{"Test-Header-Get", "Test-Header-PutPatch"},
			expectedMethods: []string{"GET", "PUT", "PATCH", "DELETE"},
		},
	}
	for _, tc := range testCases {
		tc := tc
		setupAndDoRequest(t, &tc)

		assert.Equal(t, tc.allowedHeaders, tc.actualHeaders)
		assert.Equal(t, tc.expectedMethods, tc.actualMethods)
	}
}
