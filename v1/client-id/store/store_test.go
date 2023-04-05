package store_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/SKF/go-utility/v2/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/SKF/go-enlight-middleware/v1/client-id/models"
	"github.com/SKF/go-enlight-middleware/v1/client-id/store"
)

type mockedStore struct {
	mock.Mock
}

func (m *mockedStore) GetClientID(ctx context.Context, ID string) (models.ClientID, error) {
	args := m.Called(ctx, ID)
	return args.Get(0).(models.ClientID), args.Error(1)
}

func TestCache_Simple(t *testing.T) {
	ctx := context.Background()
	expected := models.ClientID{
		Identifier: uuid.New(),
	}
	mockedStore := &mockedStore{}
	mockedStore.On("GetClientID", ctx, expected.Identifier.String()).Return(expected, nil).Once()

	cache := &store.Cache{
		Store: mockedStore,
		TTL:   1 * time.Millisecond,
	}

	for i := 0; i < 3; i++ {
		actual, err := cache.GetClientID(ctx, expected.Identifier.String())

		require.NoError(t, err)
		require.Equal(t, expected, actual)
	}

	mockedStore.AssertExpectations(t)
}

func TestCache_NotFoundPropegation(t *testing.T) {
	ctx := context.Background()
	mockedStore := &mockedStore{}
	mockedStore.On("GetClientID", ctx, mock.Anything).Return(models.ClientID{}, store.ErrNotFound)

	cache := &store.Cache{
		Store: mockedStore,
		TTL:   1 * time.Millisecond,
	}

	_, err := cache.GetClientID(ctx, "Missing")

	require.Error(t, err)
	require.True(t, errors.Is(err, store.ErrNotFound))

	mockedStore.AssertExpectations(t)
}

func TestCache_TTL(t *testing.T) {
	ctx := context.Background()
	expected := models.ClientID{
		Identifier: uuid.New(),
	}
	mockedStore := &mockedStore{}
	mockedStore.On("GetClientID", ctx, expected.Identifier.String()).Return(expected, nil).Twice()

	cache := &store.Cache{
		Store: mockedStore,
		TTL:   10 * time.Millisecond,
	}

	for _, sleep := range []time.Duration{
		cache.TTL / 10,
		cache.TTL / 2,
		cache.TTL / 2,
		cache.TTL / 10,
	} {
		actual, err := cache.GetClientID(ctx, expected.Identifier.String())
		require.NoError(t, err)
		require.Equal(t, expected, actual)

		time.Sleep(sleep)
	}

	mockedStore.AssertExpectations(t)
}

func TestCache_NoTTL(t *testing.T) {
	ctx := context.Background()
	expected := models.ClientID{
		Identifier: uuid.New(),
	}
	mockedStore := &mockedStore{}
	mockedStore.On("GetClientID", ctx, expected.Identifier.String()).Return(expected, nil).Twice()

	cache := &store.Cache{
		Store: mockedStore,
	}

	for i := 0; i < 2; i++ {
		actual, err := cache.GetClientID(ctx, expected.Identifier.String())

		require.NoError(t, err)
		require.Equal(t, expected, actual)
	}

	mockedStore.AssertExpectations(t)
}
