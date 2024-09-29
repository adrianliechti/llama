package custom

type Option func(*Client)

func WithName(val string) Option {
	return func(c *Client) {
		c.name = val
	}
}

func WithDescription(val string) Option {
	return func(c *Client) {
		c.description = val
	}
}
