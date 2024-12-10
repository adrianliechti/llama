package google

import (
	"errors"

	"google.golang.org/api/googleapi"
)

func convertError(err error) error {
	var apierr *googleapi.Error

	if errors.As(err, &apierr) {
		return errors.New(apierr.Body)
	}

	return err
}
