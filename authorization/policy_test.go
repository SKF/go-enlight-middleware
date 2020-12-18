package authorization

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	authorize_mock "github.com/SKF/go-enlight-sdk/v2/services/authorize/mock"
	proto "github.com/SKF/proto/v2/common"
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

	authorizerMock.On("IsAuthorizedWithContext", ctx, userID, policy.Action, (*proto.Origin)(nil)).Return(true, nil)

	require.NoError(t, policy.Authorize(ctx, userID, authorizerMock, request))
}
