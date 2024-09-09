package custom

type Config struct {
	url string

	model string
}

type Option func(*Config)
