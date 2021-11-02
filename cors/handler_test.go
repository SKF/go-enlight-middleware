package cors_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	cors_mw "github.com/SKF/go-enlight-middleware/cors"
)

func Test_AccessControlHeaders_MethodGet(t *testing.T) {
	options := []cors_mw.Option{}

	middleware := cors_mw.New(options...)

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
	options := []cors_mw.Option{}

	middleware := cors_mw.New(options...)

	request := httptest.NewRequest(http.MethodOptions, "/", nil)

	response := doRequest(request, middleware)
	defer response.Body.Close()

	allowHeaders := response.Header.Get("Access-Control-Allow-Headers")
	require.Equal(t, "*", allowHeaders)

	allowMethods := response.Header.Get("Access-Control-Allow-Methods")
	require.Equal(t, "*", allowMethods)

	allowOrigin := response.Header.Get("Access-Control-Allow-Origin")
	require.Equal(t, "*", allowOrigin)

	body := CorsEcho{}
	json.NewDecoder(response.Body).Decode(&body) // nolint
	require.False(t, body.Found)

	require.Equal(t, http.StatusOK, response.StatusCode)
}

type CorsEcho struct {
	Found bool
}

func doRequest(request *http.Request, mw *cors_mw.Middleware) *http.Response {
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
