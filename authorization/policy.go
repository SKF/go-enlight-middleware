package authorization

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/SKF/go-enlight-sdk/v2/services/authorize"

	proto "github.com/SKF/proto/v2/common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	custom_problems "github.com/SKF/go-enlight-middleware/authorization/problems"
)

type Policy interface {
	Authorize(ctx context.Context, userID string, authorizer AuthorizerClient, r *http.Request) error
}

type ResourceExtractor func(ctx context.Context, r *http.Request) (*proto.Origin, error)

type ActionResourcePolicy struct {
	Action            string
	ResourceExtractor ResourceExtractor
}

func (p ActionResourcePolicy) Authorize(ctx context.Context, userID string, authorizer AuthorizerClient, r *http.Request) error {
	var resource *proto.Origin

	if p.ResourceExtractor != nil {
		var err error
		if resource, err = p.ResourceExtractor(ctx, r); err != nil {
			return err
		}
	}

	ok, reason, err := authorizer.IsAuthorizedWithReasonWithContext(ctx, userID, p.Action, resource)
	if status.Code(err) == codes.Canceled {
		return context.Canceled
	} else if err != nil {
		return fmt.Errorf("unable to call IsAuthorizedWithContext: %w", err)
	}

	if !ok {
		if reason == authorize.ReasonResourceNotFound {
			return custom_problems.ResourceNotFound(resource.Id, resource.Type)
		}

		return custom_problems.Unauthorized(userID, custom_problems.PolicyViolation{
			Resource: resource.Id,
			Action:   p.Action,
		})
	}

	return nil
}

type MultiPolicy []Policy

func (policies MultiPolicy) Authorize(ctx context.Context, userID string, authorizer AuthorizerClient, r *http.Request) error {
	anyErr := false
	multiErr := custom_problems.Unauthorized(userID)

	for _, policy := range policies {
		if err := policy.Authorize(ctx, userID, authorizer, r); err != nil {
			anyErr = true

			var problem custom_problems.UnauthorizedProblem
			if errors.As(err, &problem) {
				multiErr.Violations = append(multiErr.Violations, problem.Violations...)
				continue
			}

			return err
		}
	}

	if anyErr {
		return multiErr
	}

	return nil
}
