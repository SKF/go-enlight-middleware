package recovery

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SKF/go-rest-utility/problems"
	"github.com/stretchr/testify/require"
)

func TestPanicOutputsAnInternalProblem(t *testing.T) {
	panicValue := ":/"
	middleware := New().Middleware()
	endpoint := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		panic(panicValue)
	})

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	middleware(endpoint).ServeHTTP(w, request)

	response := w.Result()
	defer response.Body.Close()

	var problem problems.BasicProblem
	err := json.NewDecoder(response.Body).Decode(&problem)
	require.NoError(t, err)

	require.Equal(t, http.StatusInternalServerError, response.StatusCode)
	require.Equal(t, "application/problem+json", response.Header.Get("Content-Type"))
	require.Equal(t, "/problems/internal-server-error", problem.ProblemType())
	require.NotContains(t, problem.Type, panicValue, "Do not leak panic information")
	require.NotContains(t, problem.Detail, panicValue, "Do not leak panic information")
}
