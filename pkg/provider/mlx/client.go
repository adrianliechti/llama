package mlx

type Client struct {
	*Completer
}

func New(url string, options ...Option) (*Client, error) {
	c, err := NewCompleter(url, options...)

	if err != nil {
		return nil, err
	}

	return &Client{
		Completer: c,
	}, nil
}
