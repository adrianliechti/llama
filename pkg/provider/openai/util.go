package openai

import (
	"errors"

	"github.com/openai/openai-go"
)

func convertError(err error) error {
	var apierr *openai.Error

	if errors.As(err, &apierr) {
		//println(string(apierr.DumpRequest(true)))  // Prints the serialized HTTP request
		//println(string(apierr.DumpResponse(true))) // Prints the serialized HTTP response
	}

	return err
}
