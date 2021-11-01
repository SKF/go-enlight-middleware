package cors_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	cors_mw "github.com/SKF/go-enlight-middleware/cors"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

func Test_Get(t *testing.T) {
	options := []cors_mw.Option{}

	middleware := cors_mw.New(options...)

	request := httptest.NewRequest(http.MethodGet, "/", nil)

	response := doRequest(request, middleware)
	defer response.Body.Close()

	allowOrigin, allowOriginExist := response.Header["Access-Control-Allow-Origin"]
	require.True(t, allowOriginExist)
	require.Len(t, allowOrigin, 1)
	require.Equal(t, "*", allowOrigin[0])

	var body CorsEcho
	json.NewDecoder(response.Body).Decode(&body) // nolint
	require.True(t, body.Found)

	require.Equal(t, http.StatusOK, response.StatusCode)
}

func Test_Options(t *testing.T) {
	options := []cors_mw.Option{}

	middleware := cors_mw.New(options...)

	request := httptest.NewRequest(http.MethodOptions, "/", nil)

	response := doRequest(request, middleware)
	defer response.Body.Close()

	allowHeaders, allowHeadersExist := response.Header["Access-Control-Allow-Headers"]
	require.True(t, allowHeadersExist)
	require.Len(t, allowHeaders, 1)
	require.Equal(t, "*", allowHeaders[0])

	allowMethods, allowMethodsExist := response.Header["Access-Control-Allow-Methods"]
	require.True(t, allowMethodsExist)
	require.Len(t, allowMethods, 1)
	require.Equal(t, "*", allowMethods[0])

	allowOrigin, allowOriginExist := response.Header["Access-Control-Allow-Origin"]
	require.True(t, allowOriginExist)
	require.Len(t, allowOrigin, 1)
	require.Equal(t, "*", allowOrigin[0])

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
