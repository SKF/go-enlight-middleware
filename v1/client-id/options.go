package clientid

import (
	"time"

	aws_arn "github.com/aws/aws-sdk-go/aws/arn"
	aws_client "github.com/aws/aws-sdk-go/aws/client"

	"github.com/SKF/go-enlight-middleware/v1/client-id/enforcement"
	"github.com/SKF/go-enlight-middleware/v1/client-id/extractor"
	"github.com/SKF/go-enlight-middleware/v1/client-id/models"
	"github.com/SKF/go-enlight-middleware/v1/client-id/store"
)

type Option func(*Middleware)

func WithStage(stage string) Option {
	return func(m *Middleware) {
		m.stage = models.Environment(stage)
	}
}

func WithRequired() Option {
	return func(m *Middleware) {
		m.enforcement = enforcement.BinaryPolicy(true)
	}
}

func WithS3Store(session aws_client.ConfigProvider, arn string) Option {
	parsedArn, _ := aws_arn.Parse(arn) //nolint:errcheck

	return WithStore(
		store.NewS3Store(session, parsedArn),
	)
}

func WithStore(s Store) Option {
	return func(m *Middleware) {
		if _, ok := s.(*store.Cache); ok {
			m.store = s
		} else if cache, ok := m.store.(*store.Cache); ok {
			cache.Store = s
		} else {
			m.store = s
		}
	}
}

func WithStoreCache() Option {
	return func(m *Middleware) {
		if _, ok := m.store.(*store.Cache); !ok {
			m.store = &store.Cache{
				Store: m.store,
				TTL:   1 * time.Hour,
			}
		}
	}
}

func WithHeaderExtractor(headers ...string) Option {
	return WithExtractor(
		extractor.HeaderExtractor(headers),
	)
}

func WithExtractor(e extractor.Extractor) Option {
	return func(m *Middleware) {
		m.extractor = e
	}
}

func WithEnforcmentPolicy(p enforcement.Policy) Option {
	return func(m *Middleware) {
		m.enforcement = p
	}
}
