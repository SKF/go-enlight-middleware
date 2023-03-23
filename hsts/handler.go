package hsts

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	middleware "github.com/SKF/go-enlight-middleware"
)

const (
	Header        string        = "Strict-Transport-Security"
	DefaultMaxAge time.Duration = 365 * 24 * time.Hour
)

type Middleware struct {
	maxAge            time.Duration
	includeSubDomains bool
	preload           bool

	cachedPolicy string
}

func New(opts ...Option) *Middleware {
	m := &Middleware{
		maxAge: DefaultMaxAge,
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

func (m *Middleware) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, span := middleware.StartSpan(r.Context(), "HSTS")

			if m.isHTTPS(r) {
				if m.cachedPolicy == "" {
					m.cachedPolicy = m.buildPolicy()
				}

				w.Header().Add(Header, m.cachedPolicy)
			}

			span.End()
			next.ServeHTTP(w, r)
		})
	}
}

func (m *Middleware) isHTTPS(r *http.Request) bool {
	if r.TLS != nil && r.TLS.HandshakeComplete {
		return true
	}

	if r.Header.Get("X-Forwarded-Proto") == "https" {
		return true
	}

	return false
}

func (m *Middleware) buildPolicy() string {
	policy := new(strings.Builder)

	policy.WriteString("max-age=")
	policy.WriteString(strconv.Itoa(int(m.maxAge.Seconds())))

	if m.includeSubDomains {
		policy.WriteString("; includeSubDomains")
	}

	if m.preload {
		policy.WriteString("; preload")
	}

	return policy.String()
}
