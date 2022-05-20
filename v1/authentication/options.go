package authentication

import "github.com/SKF/go-utility/v2/jwk"

type Option func(*Middleware)

func WithStage(stage string) Option {
	return func(m *Middleware) {
		jwk.Configure(jwk.Config{Stage: stage})

		m.ssoClient = NewSSOClient(stage)
	}
}
