package store

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/SKF/go-utility/v2/uuid"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"gopkg.in/yaml.v3"

	"github.com/SKF/go-enlight-middleware/client-id/models"
)

const (
	s3ReloadRate = 1 * time.Minute
	regionHint   = "eu-west-1"
)

type s3Store struct {
	Client s3Client

	Bucket string
	Key    string

	cacheMutex *sync.RWMutex
	lastReload time.Time
	lastETag   *string
	cache      models.ClientIDs
}

type s3Client interface {
	GetObjectWithContext(context.Context, *s3.GetObjectInput, ...request.Option) (*s3.GetObjectOutput, error)
}

func NewS3Store(cp client.ConfigProvider, arn arn.ARN) Store {
	ctx := context.Background()

	resourceParts := strings.SplitN(arn.Resource, "/", 2)
	bucket, key := resourceParts[0], resourceParts[1]

	if arn.Region == "" {
		var err error
		if arn.Region, err = s3manager.GetBucketRegion(ctx, cp, bucket, regionHint); err != nil {
			arn.Region = regionHint
		}
	}

	return &s3Store{
		Client:     s3.New(cp, aws.NewConfig().WithRegion(arn.Region)),
		Bucket:     bucket,
		Key:        key,
		cacheMutex: new(sync.RWMutex),
	}
}

func (s *s3Store) reloadCache(ctx context.Context) error {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	if time.Since(s.lastReload) <= s3ReloadRate {
		return nil
	}

	response, err := s.Client.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket:      &s.Bucket,
		Key:         &s.Key,
		IfNoneMatch: s.lastETag,
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == "NotModified" {
			s.lastReload = time.Now()
			return nil
		}

		return fmt.Errorf("unable to fetch config from s3: %w", err)
	}

	defer response.Body.Close()

	if err := yaml.NewDecoder(response.Body).Decode(&s.cache); err != nil {
		return fmt.Errorf("unable to parse config from s3: %w", err)
	}

	for identifier, cid := range s.cache {
		cid.Identifier = identifier
		s.cache[identifier] = cid
	}

	s.lastETag = response.ETag
	s.lastReload = time.Now()

	return nil
}

func (s *s3Store) GetClientID(ctx context.Context, ID string) (models.ClientID, error) {
	// default to cache if last fetch was xx seconds ago, to avoid spikes
	if time.Since(s.lastReload) > s3ReloadRate {
		if err := s.reloadCache(ctx); err != nil {
			return models.ClientID{}, fmt.Errorf("unable to reload config from s3: %w", err)
		}
	}

	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()

	cid, found := s.cache[uuid.UUID(ID)]
	if !found {
		return cid, ErrNotFound
	}

	return cid, nil
}
