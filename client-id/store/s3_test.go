package store

import (
	"bytes"
	"context"
	"io"
	"sync"
	"testing"

	"github.com/SKF/go-utility/v2/uuid"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"gopkg.in/yaml.v3"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/SKF/go-enlight-middleware/client-id/models"
)

type clientMock struct {
	mock.Mock
}

func (mock *clientMock) GetObjectWithContext(ctx context.Context, inp *s3.GetObjectInput, opt ...request.Option) (*s3.GetObjectOutput, error) {
	args := mock.Called(ctx, inp, opt)
	return args.Get(0).(*s3.GetObjectOutput), args.Error(1)
}

func TestS3GetClientID_Happy(t *testing.T) {
	var (
		expectedBucket     = "bucket"
		expectedKey        = "key"
		expectedIdentifier = uuid.UUID("2bf8888d-c379-415d-b532-b829400964f6")
	)

	b := new(bytes.Buffer)

	ids := models.ClientIDs{
		expectedIdentifier: {
			Name: "Test 1",
		},
	}

	err := yaml.NewEncoder(b).Encode(ids)
	require.NoError(t, err)

	ctx := context.Background()
	client := new(clientMock)
	expectedInput := mock.MatchedBy(func(i *s3.GetObjectInput) bool {
		return *i.Bucket == expectedBucket && *i.Key == expectedKey
	})

	client.On("GetObjectWithContext", ctx, expectedInput, mock.Anything).Return(&s3.GetObjectOutput{
		Body: io.NopCloser(b),
	}, nil).Once()

	store := s3Store{
		Client:     client,
		Bucket:     expectedBucket,
		Key:        expectedKey,
	}

	cid, err := store.GetClientID(ctx, string(expectedIdentifier))
	require.NoError(t, err)

	require.Equal(t, expectedIdentifier, cid.Identifier)
	require.Equal(t, ids[expectedIdentifier].Name, cid.Name)

	client.AssertExpectations(t)
}
