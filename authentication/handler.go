package authentication

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/mux"
	old_errors "github.com/pkg/errors"

	"github.com/SKF/go-rest-utility/problems"
	"github.com/SKF/go-utility/v2/accesstokensubcontext"
	"github.com/SKF/go-utility/v2/jwk"
	"github.com/SKF/go-utility/v2/jwt"
	"github.com/SKF/go-utility/v2/log"
	"github.com/SKF/go-utility/v2/stages"
	"github.com/SKF/go-utility/v2/useridcontext"

	jwt_go "github.com/golang-jwt/jwt/v4"
	jwt_request "github.com/golang-jwt/jwt/v4/request"

	middleware "github.com/SKF/go-enlight-middleware"
	custom_problems "github.com/SKF/go-enlight-middleware/authentication/problems"
)

type Middleware struct {
	TokenExtractor jwt_request.Extractor
	Tracer         middleware.Tracer

	unauthenticatedRoutes []*mux.Route
	userIDCache           *sync.Map // map[jwt.Subject]EnlightUserID
	ssoClient             *SSOClient
}

func New(opts ...Option) *Middleware {
	defaultStage := stages.StageProd

	jwk.Configure(jwk.Config{Stage: defaultStage})

	m := &Middleware{
		TokenExtractor: jwt_request.AuthorizationHeaderExtractor,
		Tracer:         new(middleware.OpenCensusTracer),

		unauthenticatedRoutes: []*mux.Route{},
		userIDCache:           new(sync.Map),
		ssoClient:             NewSSOClient(defaultStage),
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
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
			ctx, span := m.Tracer.StartSpan(r.Context(), "Middleware/Authentication")

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
	_, span := m.Tracer.StartSpan(ctx, "Middleware/Authentication/isAuthenticationNeeded")
	defer span.End()

	if route := mux.CurrentRoute(r); route != nil {
		for _, unauthenticatedRoute := range m.unauthenticatedRoutes {
			if route == unauthenticatedRoute {
				return false
			}
		}
	}

	return true
}

func (m *Middleware) parseFromRequest(ctx context.Context, r *http.Request) (*jwt.Token, error) {
	_, span := m.Tracer.StartSpan(ctx, "Middleware/Authentication/parseFromRequest")
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
		return nil, jwtErrorToProblem(err)
	}

	return &token, nil
}

// decorateValidRequest attatches the Cognito and Enlight UserID onto the Request Context.
func (m *Middleware) decorateValidRequest(ctx context.Context, r *http.Request, token *jwt.Token) (*http.Request, error) {
	ctx, span := m.Tracer.StartSpan(ctx, "Middleware/Authentication/decorateValidRequest")
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

	rCtx := r.Context()
	rCtx = accesstokensubcontext.NewContext(rCtx, claims.Subject)
	rCtx = useridcontext.NewContext(rCtx, userID)

	return r.WithContext(rCtx), nil
}

func jwtErrorToProblem(err error) error {
	var ve *jwt_go.ValidationError
	if errors.As(err, &ve) {
		switch {
		case ve.Errors&jwt_go.ValidationErrorMalformed != 0:
			return custom_problems.MalformedToken()
		case ve.Errors&jwt_go.ValidationErrorUnverifiable != 0:
			// JWT KeyFunc errors
			return custom_problems.UnverifiableToken()
		case ve.Errors&jwt_go.ValidationErrorSignatureInvalid != 0:
			return custom_problems.UnverifiableToken()
		case ve.Errors&jwt_go.ValidationErrorExpired != 0:
			return custom_problems.ExpiredToken()
		case ve.Errors&jwt_go.ValidationErrorNotValidYet != 0:
			return custom_problems.NotYetValidToken()
		}

		return custom_problems.InvalidToken(ve.Error())
	}

	if strings.HasPrefix(err.Error(), "token is not valid") {
		return custom_problems.InvalidToken("The provided authentication token is invalid.")
	}

	if strings.HasPrefix(err.Error(), "parse with claims failed:") ||
		strings.HasPrefix(err.Error(), "failed to validate claims:") {
		return custom_problems.InvalidToken(old_errors.Cause(err).Error())
	}

	// Will be remapped to InternalProblem by problems.WriteResponse.
	return err
}
