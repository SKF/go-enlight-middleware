package authorization

import middleware "github.com/SKF/go-enlight-middleware"

type Option func(*Middleware)

func WithAuthorizerClient(client AuthorizerClient) Option {
	return func(m *Middleware) {
		m.authorizerClient = client
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
