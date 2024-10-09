package anthropic

import (
	"errors"

	"github.com/anthropics/anthropic-sdk-go"
)

func convertError(err error) error {
	var apierr *anthropic.Error

	if errors.As(err, &apierr) {
		//println(string(apierr.DumpRequest(true)))  // Prints the serialized HTTP request
		//println(string(apierr.DumpResponse(true))) // Prints the serialized HTTP response
	}

	return err
}
