package utils

import (
	"fmt"
	"net/http"
	"strconv"

	errs "github.com/nseve/music-catalog/internal/core/errors"
)

func GetIntQueryParam(r *http.Request, key string) (*int, error) {
	param := r.URL.Query().Get(key)
	if param == "" {
		return nil, nil
	}

	val, err := strconv.Atoi(param)
	if err != nil {
		return nil, fmt.Errorf(
			"query parameter %q must be an integer (got %q): %v: %w",
			key,
			param,
			err,
			errs.ErrInvalidArgument,
		)
	}

	return &val, nil
}
