package clientid

import (
	"net/http"
	"time"

	"github.com/SKF/go-rest-utility/problems"
	"github.com/SKF/go-utility/v2/stages"

	middleware "github.com/SKF/go-enlight-middleware"
	"github.com/SKF/go-enlight-middleware/client-id/enforcement"
	"github.com/SKF/go-enlight-middleware/client-id/extractor"
	"github.com/SKF/go-enlight-middleware/client-id/models"
	custom_problems "github.com/SKF/go-enlight-middleware/client-id/problems"
	"github.com/SKF/go-enlight-middleware/client-id/store"
)

type Middleware struct {
	Tracer middleware.Tracer

	Stage models.Environment

	Extractor   extractor.Extractor
	Enforcement enforcement.Policy
	Store       store.Store
}

type (
	ClientID = models.ClientID
	Store    = store.Store
)

func New(opts ...Option) *Middleware {
	m := &Middleware{
		Tracer: new(middleware.OpenCensusTracer),

		Stage: stages.StageProd,

		Extractor:   extractor.Default,
		Enforcement: enforcement.Default,
		Store:       new(store.Default),
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

func (m *Middleware) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, span := m.Tracer.StartSpan(r.Context(), "Middleware/ClientID")

			identifier, err := m.Extractor.ExtractClientID(r)
			if enforcement := m.Enforcement.OnExtraction(ctx, err); enforcement != nil {
				problems.WriteResponse(ctx, enforcement, w, r)
				span.End()
				return
			}

			cid, err := m.Store.GetClientID(ctx, identifier)
			if enforcement := m.Enforcement.OnRetreival(ctx, err); enforcement != nil {
				problems.WriteResponse(ctx, enforcement, w, r)
				span.End()
				return
			}

			if !cid.IsEmpty() {
				err = m.validateClientID(cid)
				if enforcement := m.Enforcement.OnValidation(ctx, err); enforcement != nil {
					problems.WriteResponse(ctx, enforcement, w, r)
					span.End()
					return
				}
			}

			if err == nil {
				r = r.WithContext(
					cid.EmbedIntoContext(r.Context()),
				)
			}

			span.End()
			next.ServeHTTP(w, r)
		})
	}
}

func (m *Middleware) validateClientID(cid ClientID) error {
	if !cid.Environments.Contains(m.Stage) {
		return custom_problems.UnauthorizedClientID()
	}

	now := time.Now()

	if !cid.NotBefore.IsZero() && cid.NotBefore.After(now) {
		return custom_problems.NotYetActiveClientID(cid.NotBefore)
	}

	if !cid.Expires.IsZero() && cid.Expires.Before(now) {
		return custom_problems.ExpiredClientID(cid.Expires)
	}

	return nil
}
