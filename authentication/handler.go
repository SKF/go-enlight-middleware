package authentication

import (
	"context"
	"errors"
	"net/http"
	"sync"

	"github.com/SKF/go-rest-utility/problems"
	"github.com/SKF/go-utility/v2/accesstokensubcontext"
	"github.com/SKF/go-utility/v2/jwk"
	"github.com/SKF/go-utility/v2/jwt"
	"github.com/SKF/go-utility/v2/log"
	"github.com/SKF/go-utility/v2/useridcontext"
	"github.com/gorilla/mux"
	"go.opencensus.io/trace"

	jwt_go "github.com/dgrijalva/jwt-go"
	jwt_request "github.com/dgrijalva/jwt-go/request"

	custom_problems "github.com/SKF/go-enlight-middleware/authentication/problems"
	"github.com/SKF/go-enlight-middleware/common"
)

type Middleware struct {
	TokenExtractor jwt_request.Extractor

	unauthenticatedRoutes []*mux.Route
	userIDCache           *sync.Map // map[jwt.Subject]EnlightUserID
	ssoClient             *SSOClient
}

func ForStage(stage string) *Middleware {
	jwk.Configure(jwk.Config{Stage: stage})

	return &Middleware{
		TokenExtractor: jwt_request.AuthorizationHeaderExtractor,

		unauthenticatedRoutes: []*mux.Route{},
		userIDCache:           new(sync.Map),
		ssoClient:             NewSSOClient(stage),
	}
}

func (m *Middleware) IgnoreRoute(route *mux.Route) *Middleware {
	m.unauthenticatedRoutes = append(m.unauthenticatedRoutes, route)
	return m
}

func (m *Middleware) Middleware() func(http.Handler) http.Handler {
	if err := jwk.RefreshKeySets(); err != nil {
		log.WithError(err).Error("Unable to refresh JWKS in AuthenticationMiddleware")
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, span := common.StartSpan(r.Context(), "Middleware/Authentication")

			if m.isAuthenticationNeeded(ctx, r) {
				token, err := m.parseFromRequest(ctx, r)
				if err != nil {
					problems.WriteResponse(ctx, err, w, r)
					span.End()
					return
				}

				r, err = m.decorateValidRequest(ctx, r, token)
				if err != nil {
					problems.WriteResponse(ctx, err, w, r)
					span.End()
					return
				}
			}

			span.End()
			next.ServeHTTP(w, r)
		})
	}
}

func (m *Middleware) isAuthenticationNeeded(ctx context.Context, r *http.Request) bool {
	_, span := trace.StartSpan(ctx, "Middleware/Authentication/isAuthenticationNeeded")
	defer span.End()

	if len(m.unauthenticatedRoutes) > 0 {
		route := mux.CurrentRoute(r)
		if route == nil {
			// Default to require authentication if the current route can not be found.
			return true
		}

		for _, unauthenticatedRoute := range m.unauthenticatedRoutes {
			if route == unauthenticatedRoute {
				return false
			}
		}
	}

	return true
}

func (m *Middleware) parseFromRequest(ctx context.Context, r *http.Request) (*jwt.Token, error) {
	_, span := trace.StartSpan(ctx, "Middleware/Authentication/parseFromRequest")
	defer span.End()

	rawToken, err := m.TokenExtractor.ExtractToken(r)
	if err != nil {
		if errors.Is(err, jwt_request.ErrNoTokenInRequest) {
			return nil, custom_problems.NoToken()
		}

		return nil, err
	}

	token, err := jwt.Parse(rawToken)
	if err != nil {
		var ve jwt_go.ValidationError
		if errors.As(err, &ve) {
			if ve.Errors&jwt_go.ValidationErrorMalformed != 0 {
				return nil, custom_problems.MalformedToken()
			} else if ve.Errors&jwt_go.ValidationErrorUnverifiable != 0 {
				// Token could not be verified because of signing problems
				return nil, custom_problems.UnverifiableToken()
			} else if ve.Errors&jwt_go.ValidationErrorSignatureInvalid != 0 {
				// Signature validation failed
				return nil, custom_problems.InvalidToken("") // TODO Add details
			} else if ve.Errors&jwt_go.ValidationErrorExpired != 0 {
				return nil, custom_problems.ExpiredToken()
			} else if ve.Errors&jwt_go.ValidationErrorNotValidYet != 0 {
				return nil, custom_problems.NotYetValidToken()
			} else {
				// TODO add details
				return nil, custom_problems.InvalidToken("")
			}
		}

		// TODO Find JWKS errors

		/*
			failed to get key sets: InternalError
			parse with claims failed:
			token is not valid:
			failed to validate claims:
		*/

		return nil, err
	}

	return &token, nil
}

// decorateValidRequest attatches the Cognito and Enlight UserID onto the Request Context.
func (m *Middleware) decorateValidRequest(ctx context.Context, r *http.Request, token *jwt.Token) (*http.Request, error) {
	ctx, span := trace.StartSpan(ctx, "Middleware/Authentication/decorateValidRequest")
	defer span.End()

	var userID string

	claims := token.GetClaims()
	switch claims.TokenUse {
	case jwt.TokenUseID:
		userID = claims.EnlightUserID
	case jwt.TokenUseAccess:
		if m.userIDCache != nil {
			if cacheLine, found := m.userIDCache.Load(claims.Subject); found {
				userID = cacheLine.(string)

				break
			}
		}

		var err error
		if userID, err = m.ssoClient.getUserIDFromAccessToken(ctx, token.Raw); err != nil {
			return r, err
		}

		if m.userIDCache != nil {
			m.userIDCache.Store(claims.Subject, userID)
		}

	default:
		// Unreachable since jwt.Parse validates the TokenUse to be either "id" or "access".
		return r, errors.New("assertion failed: unreachable state of 'tokenUse' reached")
	}

	ctx = accesstokensubcontext.NewContext(ctx, claims.Subject)
	ctx = useridcontext.NewContext(ctx, userID)

	return r.WithContext(ctx), nil
}
