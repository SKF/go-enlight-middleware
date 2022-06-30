package authentication

import (
	"testing"

	"github.com/SKF/go-rest-utility/problems"
	jwt_go "github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/require"
)

func TestParseJWTErrorsAsProblems(t *testing.T) {
	err := jwt_go.NewValidationError("bad token", jwt_go.ValidationErrorExpired)

	problem := jwtErrorToProblem(err)

	pp, ok := problem.(problems.Problem)
	require.True(t, ok)

	require.Equal(t, "/problems/expired-authentication-token", pp.ProblemType())
}
