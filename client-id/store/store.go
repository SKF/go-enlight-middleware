package store

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/SKF/go-enlight-middleware/client-id/models"
)

var ErrNotFound = errors.New("client id could not be found")

type Default = LocalStore

type Store interface {
	GetClientID(ctx context.Context, ID string) (models.ClientID, error)
}

type Cache struct {
	Store
	TTL time.Duration

	cache sync.Map
}

type cacheEntry struct {
	cid     models.ClientID
	expires time.Time
}

func (s *Cache) GetClientID(ctx context.Context, ID string) (models.ClientID, error) {
	if entry, found := s.cache.Load(ID); found {
		typedEntry, ok := entry.(cacheEntry)
		if !ok {
			return models.ClientID{}, fmt.Errorf("invalid entry in store cache of type %t: %v", entry, entry)
		}

		if typedEntry.expires.After(time.Now()) {
			return typedEntry.cid, nil
		}

		s.cache.Delete(ID)
	}

	cid, err := s.Store.GetClientID(ctx, ID)
	if err != nil {
		return cid, err
	}

	s.cache.Store(ID, cacheEntry{
		cid:     cid,
		expires: time.Now().Add(s.TTL),
	})

	return cid, nil
}
