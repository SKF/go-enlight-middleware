package clientid

import (
	"context"
	"net/http"
	"time"

	"github.com/SKF/go-rest-utility/problems"
	"github.com/SKF/go-utility/v2/stages"
	"github.com/gorilla/mux"

	middleware "github.com/SKF/go-enlight-middleware"
	"github.com/SKF/go-enlight-middleware/client-id/enforcement"
	"github.com/SKF/go-enlight-middleware/client-id/extractor"
	"github.com/SKF/go-enlight-middleware/client-id/models"
	custom_problems "github.com/SKF/go-enlight-middleware/client-id/problems"
	"github.com/SKF/go-enlight-middleware/client-id/store"
)

type Middleware struct {
	Tracer middleware.Tracer

	allowedStages models.EnvironmentMask

	extractor   extractor.Extractor
	enforcement enforcement.Policy
	store       store.Store

	notMandatoryClientIDRoutes map[*mux.Route]bool
}

type (
	ClientID = models.ClientID
	Store    = store.Store
)

var FromContext = models.FromContext

// New returns a new Client ID middleware which embedds an valid client ID if
// provided in the request. Without providing any Options the client id is extracted
// from the request header "X-Client-ID", is optional and, using an empty in-memory store.
func New(opts ...Option) *Middleware {
	m := &Middleware{
		Tracer: new(middleware.OpenCensusTracer),

		allowedStages: models.Environments{stages.StageProd}.Mask(),

		extractor:   extractor.Default,
		enforcement: enforcement.Default,
		store:       new(store.Default),

		notMandatoryClientIDRoutes: map[*mux.Route]bool{},
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
			if m.isNotMandatoryClientID(ctx, r) {
				span.End()
				next.ServeHTTP(w, r)
				return
			}

			identifier, err := m.extractor.ExtractClientID(r)
			if enforcement := m.enforcement.OnExtraction(ctx, err); enforcement != nil {
				problems.WriteResponse(ctx, enforcement, w, r)
				span.End()
				return
			}

			cid, err := m.store.GetClientID(ctx, identifier)
			if enforcement := m.enforcement.OnRetrieval(ctx, err); enforcement != nil {
				problems.WriteResponse(ctx, enforcement, w, r)
				span.End()
				return
			}

			if !cid.IsEmpty() {
				err = m.validateClientID(cid)
				if enforcement := m.enforcement.OnValidation(ctx, err); enforcement != nil {
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

func (m *Middleware) IgnoreRoute(route *mux.Route) *Middleware {
	m.notMandatoryClientIDRoutes[route] = true
	return m
}

func (m *Middleware) isNotMandatoryClientID(ctx context.Context, r *http.Request) bool {
	_, span := m.Tracer.StartSpan(ctx, "Middleware/ClientID/isNotMandatoryClientID")
	defer span.End()

	return m.notMandatoryClientIDRoutes[mux.CurrentRoute(r)]
}

func (m *Middleware) validateClientID(cid ClientID) error {
	if cid.Environments.Mask().Disjoint(m.allowedStages) {
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
