package authentication

import (
	"fmt"
	"testing"

	"github.com/SKF/go-rest-utility/problems"
	jwt_go "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseJWTErrorsAsProblems(t *testing.T) {
	err := fmt.Errorf("bad token: %w", jwt_go.ErrTokenExpired)

	problem := jwtErrorToProblem(err)

	pp, ok := problem.(problems.Problem)
	require.True(t, ok)

	assert.Equal(t, "/problems/expired-authentication-token", pp.ProblemType())
}
