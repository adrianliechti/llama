package crawler

type Option func(*Tool)

func WithName(val string) Option {
	return func(t *Tool) {
		t.name = val
	}
}

func WithDescription(val string) Option {
	return func(t *Tool) {
		t.description = val
	}
}
