package authorization

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	authorize "github.com/SKF/go-authorizer/client"

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

	ok, reason, err := authorizer.IsAuthorizedWithReason(ctx, userID, p.Action, resource)

	if code := status.Code(err); code != codes.OK {
		switch code {
		case codes.Canceled:
			return context.Canceled
		case codes.DeadlineExceeded:
			return context.DeadlineExceeded
		default:
			return fmt.Errorf("unable to call IsAuthorizedWithReasonWithContext: %w", err)
		}
	}

	if !ok {
		if resource == nil {
			return custom_problems.Unauthorized(userID, custom_problems.PolicyViolation{
				Action: p.Action,
			})
		}

		if reason == authorize.ReasonResourceNotFound {
			return custom_problems.ResourceNotFound(resource.Id, resource.Type)
		}

		return custom_problems.Unauthorized(userID, custom_problems.PolicyViolation{
			Action:       p.Action,
			Resource:     resource.Id,
			ResourceType: resource.Type,
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
