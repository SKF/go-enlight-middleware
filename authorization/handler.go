package authorization

import (
	"context"
	"errors"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/SKF/go-rest-utility/problems"
	"github.com/SKF/go-utility/v2/log"
	"github.com/SKF/go-utility/v2/useridcontext"
	proto "github.com/SKF/proto/v2/common"

	middleware "github.com/SKF/go-enlight-middleware"
)

type AuthorizerClient interface {
	IsAuthorizedWithReasonWithContext(ctx context.Context, userID, action string, resource *proto.Origin) (bool, string, error)
}

type Middleware struct {
	Tracer middleware.Tracer

	authorizerClient AuthorizerClient
	policies         map[*mux.Route][]Policy
}

var (
	ErrNoAuthenticationMiddleware = errors.New("unable to extract user id from context, missing authentication middleware?")
)

func New(opts ...Option) *Middleware {
	m := &Middleware{
		Tracer: &middleware.OpenCensusTracer{},

		authorizerClient: nil,
		policies:         map[*mux.Route][]Policy{},
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

func (m *Middleware) AddPolicy(route *mux.Route, policy Policy) *Middleware {
	m.policies[route] = append(m.policies[route], policy)

	return m
}

func (m *Middleware) SetPolicy(route *mux.Route, policy Policy) *Middleware {
	m.policies[route] = []Policy{policy}

	return m
}

func (m *Middleware) Middleware() func(http.Handler) http.Handler {
	if m.authorizerClient == nil {
		log.Warning("Unable no AuthorizerClient found in Authorization middleware, disabling authorization.")
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, span := m.Tracer.StartSpan(r.Context(), "Middleware/Authorization")

			policies, found := m.findPoliciesForRequest(ctx, r)
			if found && m.authorizerClient != nil {
				userID, ok := useridcontext.FromContext(ctx)
				if !ok {
					err := ErrNoAuthenticationMiddleware

					problems.WriteResponse(ctx, err, w, r)
					span.End()
					return
				}

				for _, policy := range policies {
					if err := policy.Authorize(ctx, userID, m.authorizerClient, r); err != nil {
						problems.WriteResponse(ctx, err, w, r)
						span.End()
						return
					}
				}
			}

			span.End()
			next.ServeHTTP(w, r)
		})
	}
}

func (m *Middleware) findPoliciesForRequest(ctx context.Context, r *http.Request) ([]Policy, bool) {
	_, span := m.Tracer.StartSpan(ctx, "Middleware/Authorization/findPoliciesForRequest")
	defer span.End()

	currentRoute := mux.CurrentRoute(r)
	if currentRoute == nil {
		return nil, false
	}

	policies, found := m.policies[currentRoute]

	return policies, found
}
