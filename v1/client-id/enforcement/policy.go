package enforcement

import (
	"context"
	"errors"

	custom_problems "github.com/SKF/go-enlight-middleware/v1/client-id/problems"
	"github.com/SKF/go-enlight-middleware/v1/client-id/store"
)

// Default policy is making client id completely optional.
var Default Policy = BinaryPolicy(false)

// Policy which decides on how errors should be enforced by the middleware.
// The policy can decide if no error should be returned, or to decorate problems
// with specific comments that might help the API consumer.
type Policy interface {
	// OnExtraction is run after the identifier is extracted from the request
	OnExtraction(context.Context, error) error

	// OnRetrieval is run after searching for the identifier in the specificed store
	OnRetrieval(context.Context, error) error

	// OnValidation is run after all specified validation steps is run on the found client id
	OnValidation(context.Context, error) error
}

// BinaryPolicy will make client-id either completely required or completely optional.
type BinaryPolicy bool

func (enforced BinaryPolicy) OnExtraction(ctx context.Context, err error) error {
	if !enforced {
		return nil
	}

	return err
}

func (enforced BinaryPolicy) OnRetrieval(ctx context.Context, err error) error {
	if errors.Is(err, store.ErrNotFound) {
		if !enforced {
			return nil
		}

		return custom_problems.UnknownClientID()
	}

	return err
}

func (enforced BinaryPolicy) OnValidation(ctx context.Context, err error) error {
	if !enforced {
		return nil
	}

	return err
}
