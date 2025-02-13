package authorization

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	authorize "github.com/SKF/go-enlight-authorizer/client"
	authorize_mock "github.com/SKF/go-enlight-authorizer/mock"
	proto "github.com/SKF/proto/v2/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActionResourcePolicyWithOnlyAction(t *testing.T) {
	ctx := context.Background()
	policy := ActionResourcePolicy{
		Action:            "Action",
		ResourceExtractor: nil,
	}
	request := httptest.NewRequest(http.MethodGet, "/", nil)

	authorizerMock := authorize_mock.Create()
	authorizerMock.On("IsAuthorizedWithReasonWithContext", ctx, userID, policy.Action, (*proto.Origin)(nil)).
		Return(true, "", nil).Once()

	require.NoError(t, policy.Authorize(ctx, userID, authorizerMock, request))

	authorizerMock.AssertExpectations(t)
}

func TestActionResourcePolicyWithOnlyAction_Unauthorized(t *testing.T) {
	ctx := context.Background()
	policy := ActionResourcePolicy{
		Action:            "Action",
		ResourceExtractor: nil,
	}
	request := httptest.NewRequest(http.MethodGet, "/", nil)

	authorizerMock := authorize_mock.Create()
	authorizerMock.On("IsAuthorizedWithReasonWithContext", ctx, userID, policy.Action, (*proto.Origin)(nil)).
		Return(false, authorize.ReasonAccessDenied, nil).Once()

	err := policy.Authorize(ctx, userID, authorizerMock, request)
	require.Error(t, err)
	assert.Equal(t, reflect.TypeOf(err).String(), "problems.UnauthorizedProblem")

	authorizerMock.AssertExpectations(t)
}
