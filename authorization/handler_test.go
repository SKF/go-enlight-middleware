package authorization

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	authorize "github.com/SKF/go-enlight-authorizer/client"
	authorize_mock "github.com/SKF/go-enlight-authorizer/mock"
	"github.com/SKF/go-rest-utility/problems"
	"github.com/SKF/go-utility/v2/useridcontext"
	proto "github.com/SKF/proto/v2/common"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	userID   = "3ac91764-e57b-4a67-aa48-da99b029c04d"
	resource = &proto.Origin{
		Id:   "904dfc2a-7103-4561-aa45-6e5d317e90eb",
		Type: "node",
	}
	policy = ActionResourcePolicy{
		Action: "HIERARCHY::GET_NODE",
		ResourceExtractor: func(ctx context.Context, r *http.Request) (*proto.Origin, error) {
			return resource, nil
		},
	}
)

func setupAndDoRequest(userID string, policy Policy, middleware *Middleware) *http.Response {
	endpoint := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte(`{"message":"success"}`)) //nolint
	})

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	r := mux.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := useridcontext.NewContext(r.Context(), userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
	r.Use(middleware.Middleware())

	route := r.Handle("/", endpoint)
	middleware.SetPolicy(route, policy)

	r.ServeHTTP(w, request)

	return w.Result()
}

func TestValidAuthorizedRequest(t *testing.T) {
	authorizerMock := authorize_mock.Create()
	authorizerMock.On("IsAuthorizedWithReasonWithContext", mock.Anything, userID, policy.Action, resource).Return(true, "", nil)

	middleware := New(WithAuthorizerClient(authorizerMock))

	response := setupAndDoRequest(userID, policy, middleware)
	defer response.Body.Close()

	require.Equal(t, http.StatusOK, response.StatusCode)
}

func TestUnauthorizedRequestOnExistingResource(t *testing.T) {
	authorizerMock := authorize_mock.Create()
	authorizerMock.On("IsAuthorizedWithReasonWithContext", mock.Anything, userID, policy.Action, resource).Return(false, authorize.ReasonAccessDenied, nil)

	middleware := New(WithAuthorizerClient(authorizerMock))

	response := setupAndDoRequest(userID, policy, middleware)
	defer response.Body.Close()

	require.Equal(t, http.StatusForbidden, response.StatusCode)

	var problem problems.BasicProblem
	err := json.NewDecoder(response.Body).Decode(&problem)
	require.NoError(t, err)

	require.Equal(t, "application/problem+json", response.Header.Get("Content-Type"))
	require.Equal(t, "/problems/unauthorized-resource", problem.ProblemType())
}

func TestUnauthorizedRequestOnMissingResource(t *testing.T) {
	authorizerMock := authorize_mock.Create()
	authorizerMock.On("IsAuthorizedWithReasonWithContext", mock.Anything, userID, policy.Action, resource).Return(false, authorize.ReasonResourceNotFound, nil)

	middleware := New(WithAuthorizerClient(authorizerMock))

	response := setupAndDoRequest(userID, policy, middleware)
	defer response.Body.Close()

	require.Equal(t, http.StatusNotFound, response.StatusCode)

	var problem problems.BasicProblem
	err := json.NewDecoder(response.Body).Decode(&problem)
	require.NoError(t, err)

	require.Equal(t, "application/problem+json", response.Header.Get("Content-Type"))
	require.Equal(t, "/problems/resource-not-found", problem.ProblemType())
}
