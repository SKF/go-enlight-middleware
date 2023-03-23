package authentication

import (
	middleware "github.com/SKF/go-enlight-middleware"
	"github.com/SKF/go-utility/v2/jwk"
)

type Option func(*Middleware)

func WithStage(stage string) Option {
	return func(m *Middleware) {
		jwk.Configure(jwk.Config{Stage: stage})
	}
}

func WithDataDogTracer() Option {
	return func(m *Middleware) {
		m.Tracer = &middleware.DataDogTracer{}
	}
}

func WithOpenCensusTracer() Option {
	return func(m *Middleware) {
		m.Tracer = &middleware.OpenCensusTracer{}
	}
}
