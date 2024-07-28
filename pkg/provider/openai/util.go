package openai

import (
	"errors"

	"github.com/sashabaranov/go-openai"
)

func convertError(err error) error {
	var oaierr *openai.APIError

	if errors.As(err, &oaierr) {
		return errors.New(oaierr.Message)
	}

	return err
}
