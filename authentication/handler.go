package authentication

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/SKF/go-rest-utility/problems"
	"github.com/SKF/go-utility/v2/accesstokensubcontext"
	"github.com/SKF/go-utility/v2/impersonatercontext"
	"github.com/SKF/go-utility/v2/jwk"
	"github.com/SKF/go-utility/v2/jwt"
	"github.com/SKF/go-utility/v2/log"
	"github.com/SKF/go-utility/v2/stages"
	"github.com/SKF/go-utility/v2/useridcontext"
	jwt_go "github.com/golang-jwt/jwt/v5"
	jwt_request "github.com/golang-jwt/jwt/v5/request"
	"github.com/gorilla/mux"

	middleware "github.com/SKF/go-enlight-middleware"
	custom_problems "github.com/SKF/go-enlight-middleware/authentication/problems"
)

const (
	userIDPrefix   = "enlightUserId:"
	authorIDPrefix = "authorId:"
)

type Middleware struct {
	TokenExtractor jwt_request.Extractor
	Tracer         middleware.Tracer

	unauthenticatedRoutes []*mux.Route
}

func New(opts ...Option) *Middleware {
	defaultStage := stages.StageProd

	jwk.Configure(jwk.Config{Stage: defaultStage})

	m := &Middleware{
		TokenExtractor: jwt_request.AuthorizationHeaderExtractor,
		Tracer:         middleware.DefaultTracer,

		unauthenticatedRoutes: []*mux.Route{},
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
			ctx, span := m.Tracer.StartSpan(r.Context(), "Authentication")

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
	_, span := m.Tracer.StartSpan(ctx, "Authentication/isAuthenticationNeeded")
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
	_, span := m.Tracer.StartSpan(ctx, "Authentication/parseFromRequest")
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
	_, span := m.Tracer.StartSpan(ctx, "Authentication/decorateValidRequest")
	defer span.End()

	var userID string

	claims := token.GetClaims()
	userID, authorID := resolveUserAndAuthor(claims)

	if claims.TokenUse != jwt.TokenUseID && claims.TokenUse != jwt.TokenUseAccess {
		return r, errors.New("assertion failed: unreachable state of 'tokenUse' reached")
	}

	rCtx := r.Context()
	rCtx = accesstokensubcontext.NewContext(rCtx, claims.Subject)
	rCtx = useridcontext.NewContext(rCtx, userID)
	rCtx = impersonatercontext.NewContext(rCtx, authorID)

	return r.WithContext(rCtx), nil
}

func jwtErrorToProblem(err error) error {
	switch {
	case errors.Is(err, jwt_go.ErrTokenMalformed):
		return custom_problems.MalformedToken()
	case errors.Is(err, jwt_go.ErrTokenUnverifiable):
		// JWT KeyFunc errors
		return custom_problems.UnverifiableToken()
	case errors.Is(err, jwt_go.ErrTokenSignatureInvalid):
		return custom_problems.UnverifiableToken()
	case errors.Is(err, jwt_go.ErrTokenExpired):
		return custom_problems.ExpiredToken()
	case errors.Is(err, jwt_go.ErrTokenNotValidYet):
		return custom_problems.NotYetValidToken()
	}

	if strings.HasPrefix(err.Error(), "token is not valid") {
		return custom_problems.InvalidToken("The provided authentication token is invalid.")
	}

	if strings.HasPrefix(err.Error(), "parse with claims failed:") ||
		strings.HasPrefix(err.Error(), "failed to validate claims:") {
		return custom_problems.InvalidToken(errors.Unwrap(err).Error())
	}

	// Will be remapped to InternalProblem by problems.WriteResponse.
	return err
}

// userID is the Enlight User ID of the authenticated/impersonated user.
// authorID is the Enlight User ID of the authenticated user who creates the token.
// If token is generated for impersonation, author indicates the admin user who wants to impersonate.
// If it is a normal token, authorID and userID are the same.
// We added these two fields to all the tokens to make sure that it will be consistent between the services.
func resolveUserAndAuthor(claims jwt.Claims) (userID string, authorID string) {
	cognitoGroups := claims.CognitoGroups

	for _, group := range cognitoGroups {
		if strings.HasPrefix(group, userIDPrefix) {
			if len(group) == len(userIDPrefix) { // nothing after the prefix
				continue
			}

			result := group[len(userIDPrefix):]

			userID = result
		}

		if strings.HasPrefix(group, authorIDPrefix) {
			if len(group) == len(authorIDPrefix) { // nothing after the prefix
				continue
			}

			result := group[len(authorIDPrefix):]

			authorID = result
		}
	}

	return
}
