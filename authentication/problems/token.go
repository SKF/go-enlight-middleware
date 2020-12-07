package problems

import (
	"net/http"

	"github.com/SKF/go-rest-utility/problems"
)

type NoTokenProblem struct {
	problems.BasicProblem
}

func NoToken() NoTokenProblem {
	return NoTokenProblem{
		BasicProblem: problems.BasicProblem{
			Type:   "/problems/authentication-token-required",
			Title:  "Authentication is Required.",
			Status: http.StatusUnauthorized,
			Detail: `The requested endpoint requires authentication using a bearer token. This should be provided through the "Authorization" HTTP header.`,
		},
	}
}

type MalformedTokenProblem struct {
	problems.BasicProblem
}

func MalformedToken() MalformedTokenProblem {
	return MalformedTokenProblem{
		BasicProblem: problems.BasicProblem{
			Type:   "/problems/malformed-authentication-token",
			Title:  "The provided authentication token is malformed.",
			Status: http.StatusBadRequest,
			Detail: "The authentication token must be a valid JWT token in Base64.",
		},
	}
}

type UnverifiableTokenProblem struct {
	problems.BasicProblem
}

func UnverifiableToken() UnverifiableTokenProblem {
	return UnverifiableTokenProblem{
		BasicProblem: problems.BasicProblem{
			Type:   "/problems/unverifiable-authentication-token",
			Title:  "Unable to verify the JWT signature.",
			Status: http.StatusBadRequest,
			Detail: "TODO.",
		},
	}
}

type ExpiredTokenProblem struct {
	problems.BasicProblem
}

func ExpiredToken() ExpiredTokenProblem {
	return ExpiredTokenProblem{
		BasicProblem: problems.BasicProblem{
			Type:   "/problems/expired-authentication-token",
			Title:  "Provided Authentication token has expired.",
			Status: http.StatusUnauthorized,
			Detail: "An access token or identity token is only valid for 60 minutes. TODO",
		},
	}
}

type NotYetValidTokenProblem struct {
	problems.BasicProblem
}

func NotYetValidToken() NotYetValidTokenProblem {
	return NotYetValidTokenProblem{
		BasicProblem: problems.BasicProblem{
			Type:   "/problems/not-yet-valid-authentication-token",
			Title:  "Provided Authentication token is not yet valid.",
			Status: http.StatusUnauthorized,
			Detail: "The validated token should not be used before X.",
		},
	}
}

type InvalidTokenProblem struct {
	problems.BasicProblem
}

func InvalidToken(detail string) InvalidTokenProblem {
	return InvalidTokenProblem{
		BasicProblem: problems.BasicProblem{
			Type:   "/problems/invalid-authentication-token",
			Title:  "Your authentication token didn't validate.",
			Status: http.StatusUnauthorized,
			Detail: detail,
		},
	}
}
