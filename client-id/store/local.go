package store

import (
	"context"

	"github.com/SKF/go-enlight-middleware/client-id/models"
)

type LocalStore map[string]models.ClientID

func NewLocal() LocalStore {
	return LocalStore{}
}

func (s LocalStore) Add(cid models.ClientID) LocalStore {
	s[string(cid.Identifier)] = cid

	return s
}

func (s LocalStore) GetClientID(ctx context.Context, ID string) (models.ClientID, error) {
	cid, ok := s[ID]
	if !ok {
		return models.ClientID{}, ErrNotFound
	}

	return cid, nil
}
