package clientid

import (
	"time"

	aws_arn "github.com/aws/aws-sdk-go/aws/arn"

	"github.com/SKF/go-enlight-middleware/client-id/enforcement"
	"github.com/SKF/go-enlight-middleware/client-id/extractor"
	"github.com/SKF/go-enlight-middleware/client-id/models"
	"github.com/SKF/go-enlight-middleware/client-id/store"
)

type Option func(*Middleware)

func WithStage(stage string) Option {
	return func(m *Middleware) {
		m.Stage = models.Environment(stage)
	}
}

func WithRequired() Option {
	return func(m *Middleware) {
		m.Enforcement = enforcement.BinaryPolicy(true)
	}
}

func WithS3Store(arn string) Option {
	parsedArn, _ := aws_arn.Parse(arn) //nolint:errcheck

	return WithStore(
		store.NewS3Store(parsedArn),
	)
}

func WithStore(s Store) Option {
	return func(m *Middleware) {
		if _, ok := s.(*store.Cache); ok {
			m.Store = s
		} else if cache, ok := m.Store.(*store.Cache); ok {
			cache.Store = s
		} else {
			m.Store = s
		}
	}
}

func WithStoreCache() Option {
	return func(m *Middleware) {
		if _, ok := m.Store.(*store.Cache); !ok {
			m.Store = &store.Cache{
				Store: m.Store,
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
		m.Extractor = e
	}
}
