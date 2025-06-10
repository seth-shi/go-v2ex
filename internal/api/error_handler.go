package api

import (
	"context"
	"errors"
	"fmt"

	"github.com/seth-shi/go-v2ex/internal/model/response"
)

func errorWrapper(prefix string, err error) error {

	if err == nil {
		return nil
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("%s%s", prefix, response.ErrRequestTimeout.Error())
	}

	return fmt.Errorf("%s:%s", prefix, err)
}
