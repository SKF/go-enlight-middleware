package authorization

import (
	"context"

	proto "github.com/SKF/proto/v2/common"
)

type Option func(*Middleware)

type authorizerAdapter struct {
	DeprecatedAuthorizerClient
}

func (a *authorizerAdapter) IsAuthorizedWithReason(ctx context.Context, userID, action string, resource *proto.Origin) (bool, string, error) {
	return a.IsAuthorizedWithReasonWithContext(ctx, userID, action, resource)
}

func WithAuthorizerClient(client DeprecatedAuthorizerClient) Option {
	return func(m *Middleware) {
		m.authorizerClient = &authorizerAdapter{client}
	}
}

func WithAuthorizer(client AuthorizerClient) Option {
	return func(m *Middleware) {
		m.authorizerClient = client
	}
}
