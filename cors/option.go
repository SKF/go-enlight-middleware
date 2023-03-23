package cors

import middleware "github.com/SKF/go-enlight-middleware"

type Option func(*Middleware)

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
