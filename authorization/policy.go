package authorization

import (
	"context"
	"errors"
	"fmt"
	"net/http"

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
	resource, err := p.ResourceExtractor(ctx, r)
	if err != nil {
		return fmt.Errorf("unable to extract authable resource from request: %w", err)
	}

	if ok, err := authorizer.IsAuthorizedWithContext(ctx, userID, p.Action, resource); err != nil {
		switch status.Code(err) {
		case codes.Canceled:
			return context.Canceled
		}

		return fmt.Errorf("unable to call IsAuthorizedWithContext: %w", err)
	} else if !ok {
		if _, err := authorizer.GetResourceWithContext(ctx, resource.Id, resource.Type); err != nil {
			switch status.Code(err) {
			case codes.Canceled:
				return context.Canceled
			case codes.NotFound:
				return custom_problems.ResourceNotFound(resource.Id, resource.Type)
			}

			return fmt.Errorf("unable to call GetResourceWithContext: %w", err)
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
