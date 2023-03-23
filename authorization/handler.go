package authorization

import (
	"context"
	"errors"
	"net/http"

	middleware "github.com/SKF/go-enlight-middleware"
	"github.com/SKF/go-rest-utility/problems"
	"github.com/SKF/go-utility/v2/log"
	"github.com/SKF/go-utility/v2/useridcontext"
	proto "github.com/SKF/proto/v2/common"
	"github.com/gorilla/mux"
)

type AuthorizerClient interface {
	IsAuthorizedWithReasonWithContext(ctx context.Context, userID, action string, resource *proto.Origin) (bool, string, error)
}

type Middleware struct {
	authorizerClient AuthorizerClient
	policies         map[*mux.Route]Policy
}

var (
	ErrNoAuthenticationMiddleware = errors.New("unable to extract user id from context, missing authentication middleware?")
)

func New(opts ...Option) *Middleware {
	m := &Middleware{
		authorizerClient: nil,
		policies:         map[*mux.Route]Policy{},
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

func (m *Middleware) SetPolicy(route *mux.Route, policy Policy) *Middleware {
	m.policies[route] = policy

	return m
}

func (m *Middleware) Middleware() func(http.Handler) http.Handler {
	if m.authorizerClient == nil {
		log.Warning("Unable no AuthorizerClient found in Authorization middleware, disabling authorization.")
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, span := middleware.StartSpan(r.Context(), "Authorization")

			policy, found := m.findPolicyForRequest(ctx, r)
			if found && m.authorizerClient != nil {
				userID, ok := useridcontext.FromContext(ctx)
				if !ok {
					err := ErrNoAuthenticationMiddleware

					problems.WriteResponse(ctx, err, w, r)
					span.End()
					return
				}

				if err := policy.Authorize(ctx, userID, m.authorizerClient, r); err != nil {
					if !errors.Is(err, context.Canceled) {
						problems.WriteResponse(ctx, err, w, r)
					}

					span.End()
					return
				}
			}

			span.End()
			next.ServeHTTP(w, r)
		})
	}
}

func (m *Middleware) findPolicyForRequest(ctx context.Context, r *http.Request) (Policy, bool) {
	_, span := middleware.StartSpan(ctx, "Authorization/findPolicyForRequest")
	defer span.End()

	currentRoute := mux.CurrentRoute(r)
	if currentRoute == nil {
		return nil, false
	}

	policy, found := m.policies[currentRoute]

	return policy, found
}
