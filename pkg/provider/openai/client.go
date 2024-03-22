package openai

type Client struct {
	*Embedder
	*Completer
	*Synthesizer
	*Transcriber
}

func New(options ...Option) (*Client, error) {
	e, err := NewEmbedder(options...)

	if err != nil {
		return nil, err
	}

	c, err := NewCompleter(options...)

	if err != nil {
		return nil, err
	}

	s, err := NewSynthesizer(options...)

	if err != nil {
		return nil, err
	}

	t, err := NewTranscriber(options...)

	if err != nil {
		return nil, err
	}

	return &Client{
		Embedder:    e,
		Completer:   c,
		Synthesizer: s,
		Transcriber: t,
	}, nil
}
