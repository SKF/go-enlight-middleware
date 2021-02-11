package store

import (
	"context"

	"github.com/SKF/go-enlight-middleware/client-id/models"
)

type localStore map[string]models.ClientID

func NewLocal() Store {
	return localStore{}
}

func (s localStore) GetClientID(ctx context.Context, ID string) (models.ClientID, error) {
	cid, ok := s[ID]
	if !ok {
		return models.ClientID{}, ErrNotFound
	}

	return cid, nil
}
