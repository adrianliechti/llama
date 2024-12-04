package bedrock

type Config struct {
	model string

	region string
}

type Option func(*Config)

func WithRegion(region string) Option {
	return func(c *Config) {
		c.region = region
	}
}
