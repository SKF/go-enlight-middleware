package authentication_test

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/SKF/go-utility/v2/jwk"
	"github.com/SKF/go-utility/v2/jwt"
	"github.com/SKF/go-utility/v2/useridcontext"
	jwt_request "github.com/golang-jwt/jwt/v5/request"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v2/jwa"
	ljwk "github.com/lestrrat-go/jwx/v2/jwk"
	ljwt "github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	middleware "github.com/SKF/go-enlight-middleware"
	"github.com/SKF/go-enlight-middleware/authentication"
	problems "github.com/SKF/go-enlight-middleware/authentication/problems"
)

// These tests relies on global variables in packages and can't be run in
// parallel.

func createKey(t *testing.T) (ljwk.Key, ljwk.Set) {
	// Create an RSA keypair
	valid, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// Create a JWK key from the RSA keypair
	validKey, err := ljwk.FromRaw(valid)
	require.NoError(t, err)

	// Adds fields expected by our packages
	// The "kid" needs to be different in each test to ensure
	// the library refetches the keyset.
	validKey.Set(ljwk.KeyIDKey, t.Name())      //nolint:errcheck
	validKey.Set(ljwk.AlgorithmKey, jwa.RS256) //nolint:errcheck
	validKey.Set(ljwk.KeyUsageKey, "sig")      //nolint:errcheck

	// Cast to a private key
	validJWTKey, ok := validKey.(ljwk.RSAPrivateKey)
	require.True(t, ok)

	// Create a JWKS
	validSet := ljwk.NewSet()

	err = validSet.AddKey(validJWTKey)
	require.NoError(t, err)

	return validKey, validSet
}

func createJWKSServer(t *testing.T, set ljwk.Set) *httptest.Server {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// The server returns the public keyset
		public, err := ljwk.PublicSetOf(set)
		require.NoError(t, err)
		require.NoError(t, json.NewEncoder(w).Encode(public))
	}))

	jwk.KeySetURL = s.URL

	return s
}

func createSignedToken(t *testing.T, key ljwk.Key, values map[string]any) []byte {
	// Create a token and add fields for an access token
	token := ljwt.New()

	for key, value := range values {
		token.Set(key, value) //nolint:errcheck
	}

	// Signed it using the private key
	signed, err := ljwt.Sign(token, ljwt.WithKey(jwa.RS256, key))
	require.NoError(t, err)

	return signed
}

//nolint:bodyclose
func Test_Middleware_AccesToken(t *testing.T) {
	userID := uuid.New().String()

	validKey, validSet := createKey(t)

	s := createJWKSServer(t, validSet)
	defer s.Close()

	h := (&authentication.Middleware{
		TokenExtractor: jwt_request.AuthorizationHeaderExtractor,
		Tracer:         middleware.DefaultTracer,
	}).Middleware()

	t.Run("Valid access token", func(t *testing.T) {
		signed := createSignedToken(t, validKey, map[string]any{
			"token_use": jwt.TokenUseAccess,
			"username":  "a.b@example.com",
			"cognito:groups": []string{
				fmt.Sprintf("enlightUserId:%s", userID),
			},
		})

		f := h(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			contextUserID, ok := useridcontext.FromContext(r.Context())
			require.True(t, ok)

			assert.Equal(t, userID, contextUserID)
		}))

		r, err := http.NewRequest("GET", "/", nil)
		require.NoError(t, err)

		r.Header.Add("Authorization", string(signed))

		w := httptest.NewRecorder()

		f.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	})
	t.Run("Access token missing username", func(t *testing.T) {
		signed := createSignedToken(t, validKey, map[string]any{
			"token_use": jwt.TokenUseAccess,
			"cognito:groups": []string{
				fmt.Sprintf("enlightUserId:%s", userID),
			},
		})

		f := h(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("Test should not end up here")
		}))

		r, err := http.NewRequest("GET", "/", nil)
		require.NoError(t, err)

		r.Header.Add("Authorization", string(signed))

		w := httptest.NewRecorder()

		f.ServeHTTP(w, r)

		assert.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)

		var problem problems.InvalidTokenProblem
		err = json.NewDecoder(w.Result().Body).Decode(&problem)
		require.NoError(t, err)

		assert.Equal(t, problems.InvalidToken("").Type, problem.Type)
	})
	t.Run("Valid identity token", func(t *testing.T) {
		var (
			userID    = uuid.New().String()
			companyID = uuid.New().String()
		)

		signed := createSignedToken(t, validKey, map[string]any{
			"token_use":     jwt.TokenUseID,
			"enlightUserId": userID,
			"cognito:groups": []string{
				fmt.Sprintf("enlightUserId:%s", userID),
			},
			"enlightCompanyId": companyID,
			"enlightAccess":    "",
			"enlightRoles":     "hierarchy_admin",
			"enlightEmail":     "a.b@example.com",
		})

		f := h(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			contextUserID, ok := useridcontext.FromContext(r.Context())
			require.True(t, ok)

			assert.Equal(t, userID, contextUserID)
		}))

		r, err := http.NewRequest("GET", "/", nil)
		require.NoError(t, err)

		r.Header.Add("Authorization", string(signed))

		w := httptest.NewRecorder()

		f.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	})
	t.Run("Token valid in future", func(t *testing.T) {
		userID := uuid.New().String()

		signed := createSignedToken(t, validKey, map[string]any{
			ljwt.NotBeforeKey: time.Now().UTC().Add(1 * time.Hour),
			"token_use":       jwt.TokenUseAccess,
			"username":        "a.b@example.com",
			"cognito:groups": []string{
				fmt.Sprintf("enlightUserId:%s", userID),
			},
		})

		f := h(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("Test should not end up here")
		}))

		r, err := http.NewRequest("GET", "/", nil)
		require.NoError(t, err)

		r.Header.Add("Authorization", string(signed))

		w := httptest.NewRecorder()

		f.ServeHTTP(w, r)

		assert.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)

		var problem problems.InvalidTokenProblem
		err = json.NewDecoder(w.Result().Body).Decode(&problem)
		require.NoError(t, err)

		assert.Equal(t, problems.InvalidToken("").Type, problem.Type)
	})
	t.Run("Token has expired", func(t *testing.T) {
		userID := uuid.New().String()

		signed := createSignedToken(t, validKey, map[string]any{
			ljwt.ExpirationKey: time.Now().UTC().Add(-1 * time.Hour),
			"token_use":        jwt.TokenUseAccess,
			"username":         "a.b@example.com",
			"cognito:groups": []string{
				fmt.Sprintf("enlightUserId:%s", userID),
			},
		})

		f := h(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("Test should not end up here")
		}))

		r, err := http.NewRequest("GET", "/", nil)
		require.NoError(t, err)

		r.Header.Add("Authorization", string(signed))

		w := httptest.NewRecorder()

		f.ServeHTTP(w, r)

		assert.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)

		var problem problems.InvalidTokenProblem
		err = json.NewDecoder(w.Result().Body).Decode(&problem)
		require.NoError(t, err)

		assert.Equal(t, problems.InvalidToken("").Type, problem.Type)
	})
	t.Run("Token key is unknown", func(t *testing.T) {
		userID := uuid.New().String()

		unknown, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		unknownKey, err := ljwk.FromRaw(unknown)
		require.NoError(t, err)

		unknownKey.Set(ljwk.KeyIDKey, "unknown")     //nolint:errcheck
		unknownKey.Set(ljwk.AlgorithmKey, jwa.RS256) //nolint:errcheck
		unknownKey.Set(ljwk.KeyUsageKey, "sig")      //nolint:errcheck

		signed := createSignedToken(t, unknownKey, map[string]any{
			"token_use": jwt.TokenUseAccess,
			"username":  "a.b@example.com",
			"cognito:groups": []string{
				fmt.Sprintf("enlightUserId:%s", userID),
			},
		})

		f := h(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("Test should not end up here")
		}))

		r, err := http.NewRequest("GET", "/", nil)
		require.NoError(t, err)

		r.Header.Add("Authorization", string(signed))

		w := httptest.NewRecorder()

		f.ServeHTTP(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)

		var problem problems.UnverifiableTokenProblem
		err = json.NewDecoder(w.Result().Body).Decode(&problem)
		require.NoError(t, err)

		assert.Equal(t, problems.UnverifiableToken().Type, problem.Type)
	})
	t.Run("Token is malformed", func(t *testing.T) {
		f := h(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("The test should not end up here")
		}))

		r, err := http.NewRequest("GET", "/", nil)
		require.NoError(t, err)

		r.Header.Add("Authorization", "foobar")

		w := httptest.NewRecorder()

		f.ServeHTTP(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)

		var problem problems.MalformedTokenProblem
		err = json.NewDecoder(w.Result().Body).Decode(&problem)
		require.NoError(t, err)

		assert.Equal(t, problems.MalformedToken().Type, problem.Type)
	})
}
