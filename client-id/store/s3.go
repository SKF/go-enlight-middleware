package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/SKF/go-utility/v2/uuid"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	s3manager "github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"

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
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}

func NewS3Store(awsCfg aws.Config, arn arn.ARN) Store {
	ctx := context.Background()

	resourceParts := strings.SplitN(arn.Resource, "/", 2) //nolint:gomnd
	bucket, key := resourceParts[0], resourceParts[1]

	s3Client := s3.NewFromConfig(awsCfg)

	if arn.Region == "" {
		var err error
		if arn.Region, err = s3manager.GetBucketRegion(ctx, s3Client, bucket); err != nil {
			arn.Region = regionHint
		}
	}

	awsCfg.Region = arn.Region

	return &s3Store{
		Client:     s3.NewFromConfig(awsCfg),
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

	response, err := s.Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket:      &s.Bucket,
		Key:         &s.Key,
		IfNoneMatch: s.lastETag,
	})
	if err != nil {

		var ae smithy.APIError
		if errors.As(err, &ae) && ae.ErrorCode() == "NotModified" {
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
