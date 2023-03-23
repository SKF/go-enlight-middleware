package hsts

import (
	"time"

	middleware "github.com/SKF/go-enlight-middleware"
)

const (
	oneYear  = 365 * 24 * time.Hour
	twoYears = 2 * oneYear
)

type Option func(*Middleware)

func WithMaxAge(age time.Duration) Option {
	return func(m *Middleware) {
		m.maxAge = age
	}
}

func WithIncludeSubDomains() Option {
	return func(m *Middleware) {
		m.includeSubDomains = true
	}
}

func WithPreload() Option {
	return func(m *Middleware) {
		// HSTS Preload recommendations from https://hstspreload.org/
		if m.maxAge < oneYear {
			m.maxAge = twoYears
		}

		m.includeSubDomains = true
		m.preload = true
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
