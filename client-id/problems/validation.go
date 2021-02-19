package problems

import (
	"net/http"
	"time"

	"github.com/SKF/go-rest-utility/problems"
)

type UnauthorizedClientIDProblem struct {
	problems.BasicProblem
}

func UnauthorizedClientID() UnauthorizedClientIDProblem {
	return UnauthorizedClientIDProblem{
		BasicProblem: problems.BasicProblem{
			Type:   "/problems/unauthorized-client-id",
			Title:  "The provided client id is not authorized to access the resource.",
			Status: http.StatusForbidden,
		},
	}
}

type NotYetActiveClientIDProblem struct {
	problems.BasicProblem
	Activation time.Time `json:"activation"`
}

func NotYetActiveClientID(notBefore time.Time) NotYetActiveClientIDProblem {
	return NotYetActiveClientIDProblem{
		BasicProblem: problems.BasicProblem{
			Type:   "/problems/not-yet-active-client-id",
			Title:  "Provided client ID is not yet active.",
			Status: http.StatusForbidden,
			Detail: "The provided client id is valid, but is not yet allowed to be used",
		},
		Activation: notBefore,
	}
}

type ExpiredClientIDProblem struct {
	problems.BasicProblem
	ExpiredAt time.Time `json:"expired_at"`
}

func ExpiredClientID(expiredAt time.Time) ExpiredClientIDProblem {
	return ExpiredClientIDProblem{
		BasicProblem: problems.BasicProblem{
			Type:   "/problems/expired-client-id",
			Title:  "Provided client ID has expired.",
			Status: http.StatusForbidden,
			Detail: "The client id has expired, please contact support if an extension is needed",
		},
		ExpiredAt: expiredAt,
	}
}
