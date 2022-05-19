package authorization

type Option func(*Middleware)

func WithAuthorizerClient(client AuthorizerClient) Option {
	return func(m *Middleware) {
		m.authorizerClient = client
	}
}
