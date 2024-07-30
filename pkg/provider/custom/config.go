package custom

type Config struct {
	url string

	model string
}

type Option func(*Config)

func WithModel(model string) Option {
	return func(c *Config) {
		c.model = model
	}
}
