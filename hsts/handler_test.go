package hsts_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/SKF/go-enlight-middleware/hsts"
)

func testHSTSMiddleware(t *testing.T, middleware *hsts.Middleware, server func(http.Handler) *httptest.Server, requestModifier func(*http.Request)) *http.Response {
	t.Helper()

	endpoint := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "OK")
	})

	srv := server(middleware.Middleware()(endpoint))
	defer srv.Close()

	r, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	require.NoError(t, err)

	if requestModifier != nil {
		requestModifier(r)
	}

	response, err := srv.Client().Do(r)
	require.NoError(t, err)

	return response
}

func TestHTTPRequest(t *testing.T) {
	middleware := hsts.New()

	response := testHSTSMiddleware(t, middleware, httptest.NewServer, nil)
	defer response.Body.Close()

	require.Equal(t, http.StatusOK, response.StatusCode)
	require.NotContains(t, response.Header, hsts.Header)
}

func TestForwardedHTTPSRequestOverHTTP(t *testing.T) {
	middleware := hsts.New()

	response := testHSTSMiddleware(t, middleware, httptest.NewServer, func(r *http.Request) {
		r.Header.Add("X-Forwarded-Proto", "https")
	})
	defer response.Body.Close()

	require.Equal(t, http.StatusOK, response.StatusCode)
	require.Equal(t, "max-age=31536000", response.Header.Get(hsts.Header))
}

func TestHTTPSRequest(t *testing.T) {
	middleware := hsts.New()

	response := testHSTSMiddleware(t, middleware, httptest.NewTLSServer, nil)
	defer response.Body.Close()

	require.Equal(t, http.StatusOK, response.StatusCode)
	require.Equal(t, "max-age=31536000", response.Header.Get(hsts.Header))
}

func TestWithCustomMaxAge(t *testing.T) {
	middleware := hsts.New(
		hsts.WithMaxAge(24 * time.Hour),
	)

	response := testHSTSMiddleware(t, middleware, httptest.NewTLSServer, nil)
	defer response.Body.Close()

	require.Equal(t, http.StatusOK, response.StatusCode)
	require.Equal(t, "max-age=86400", response.Header.Get(hsts.Header))
}

func TestWithIncludeSubDomains(t *testing.T) {
	middleware := hsts.New(
		hsts.WithMaxAge(hsts.DefaultMaxAge),
		hsts.WithIncludeSubDomains(),
	)

	response := testHSTSMiddleware(t, middleware, httptest.NewTLSServer, nil)
	defer response.Body.Close()

	require.Equal(t, http.StatusOK, response.StatusCode)
	require.Equal(t, "max-age=31536000; includeSubDomains", response.Header.Get(hsts.Header))
}

func TestWithPreload(t *testing.T) {
	middleware := hsts.New(
		hsts.WithMaxAge(24*time.Hour),
		hsts.WithPreload(),
	)

	response := testHSTSMiddleware(t, middleware, httptest.NewTLSServer, nil)
	defer response.Body.Close()

	require.Equal(t, http.StatusOK, response.StatusCode)
	require.Equal(t, "max-age=63072000; includeSubDomains; preload", response.Header.Get(hsts.Header))
}
