package spandecorator

type Option func(*Middleware)

func WithBody() Option {
	return func(m *Middleware) {
		m.withBody = true
	}
}
