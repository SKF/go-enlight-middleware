package cors_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	"github.com/SKF/go-enlight-middleware/v1/cors"
)

func Test_AccessControlHeaders_MethodGet(t *testing.T) {
	middleware := cors.New()

	request := httptest.NewRequest(http.MethodGet, "/", nil)

	response := doRequest(request, middleware)
	defer response.Body.Close()

	allowOrigin := response.Header.Get("Access-Control-Allow-Origin")
	require.Equal(t, "*", allowOrigin)

	var body CorsEcho
	json.NewDecoder(response.Body).Decode(&body) // nolint
	require.True(t, body.Found)

	require.Equal(t, http.StatusOK, response.StatusCode)
}

func Test_AccessControlHeaders_MethodOptions(t *testing.T) {
	middleware := cors.New()

	request := httptest.NewRequest(http.MethodOptions, "/", nil)

	response := doRequest(request, middleware)
	defer response.Body.Close()

	allowHeaders := response.Header.Get("Access-Control-Allow-Headers")
	require.Equal(t, "*", allowHeaders)

	allowMethods := response.Header.Get("Access-Control-Allow-Methods")
	require.Equal(t, "*", allowMethods)

	allowOrigin := response.Header.Get("Access-Control-Allow-Origin")
	require.Equal(t, "*", allowOrigin)

	require.Equal(t, int64(-1), response.ContentLength)

	require.Equal(t, http.StatusOK, response.StatusCode)
}

type CorsEcho struct {
	Found bool
}

func doRequest(request *http.Request, mw *cors.Middleware) *http.Response {
	endpoint := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(CorsEcho{ // nolint
			Found: true,
		})
	})

	w := httptest.NewRecorder()

	r := mux.NewRouter()
	r.Use(mw.Middleware())
	r.Handle("/", endpoint)

	r.ServeHTTP(w, request)

	return w.Result()
}
