package problems

import (
	"fmt"
	"net/http"

	"github.com/SKF/go-rest-utility/problems"
)

type ResourceNotFoundProblem struct {
	problems.BasicProblem
	Resource     string `json:"resource,omitempty"`
	ResourceType string `json:"resource_type,omitempty"`
}

func ResourceNotFound(resource, resourceType string) ResourceNotFoundProblem {
	return ResourceNotFoundProblem{
		BasicProblem: problems.BasicProblem{
			Type:   "/problems/resource-not-found",
			Title:  "The requested resource could not be found.",
			Status: http.StatusNotFound,
			Detail: fmt.Sprintf(`The %s "%s" might have been deleted.`, resourceType, resource),
		},
		Resource:     resource,
		ResourceType: resourceType,
	}
}
