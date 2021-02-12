package problems

import (
	"net/http"

	"github.com/SKF/go-rest-utility/problems"
)

type NoClientIDProblem struct {
	problems.BasicProblem
}

func NoClientID(detail string) NoClientIDProblem {
	return NoClientIDProblem{
		BasicProblem: problems.BasicProblem{
			Type:   "/problems/missing-client-id",
			Title:  "Client ID is Required.",
			Status: http.StatusUnauthorized,
			Detail: detail,
		},
	}
}

type UnknownClientIDProblem struct {
	problems.BasicProblem
}

func UnknownClientID() UnknownClientIDProblem {
	return UnknownClientIDProblem{
		BasicProblem: problems.BasicProblem{
			Type:   "/problems/unknown-client-id",
			Title:  "The provided client ID is unknown.",
			Status: http.StatusUnauthorized,
		},
	}
}
