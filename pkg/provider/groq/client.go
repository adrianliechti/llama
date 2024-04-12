package groq

type Client struct {
	*Completer
}

func New(options ...Option) (*Client, error) {
	c, err := NewCompleter(options...)

	if err != nil {
		return nil, err
	}

	return &Client{
		Completer: c,
	}, nil
}
