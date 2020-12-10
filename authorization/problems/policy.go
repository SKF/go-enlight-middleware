package problems

import (
	"net/http"

	"github.com/SKF/go-rest-utility/problems"
)

type UnauthorizedProblem struct {
	problems.BasicProblem
	User       string            `json:"user_id,omitempty"`
	Violations []PolicyViolation `json:"violations,omitempty"`
}

type PolicyViolation struct {
	Resource string `json:"resource,omitempty"`
	Action   string `json:"action,omitempty"`
}

func Unauthorized(userID string, violations ...PolicyViolation) UnauthorizedProblem {
	return UnauthorizedProblem{
		BasicProblem: problems.BasicProblem{
			Type:   "/problems/unauthorized-resource",
			Title:  "The request to access the resource was denied.",
			Status: http.StatusForbidden,
			Detail: `Your user requires more access to be able to access this resource.`,
		},
		User:       userID,
		Violations: violations,
	}
}
