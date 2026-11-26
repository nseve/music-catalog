package utils

import (
	"fmt"
	"net/http"
	"strconv"

	errs "github.com/nseve/music-catalog/internal/core/errors"
)

func GetIntPathValue(r *http.Request, key string) (int, error) {
	pathValue := r.PathValue(key)
	if pathValue == "" {
		return 0, fmt.Errorf(
			"path parameter %q is missing: %w",
			key,
			errs.ErrInvalidArgument,
		)
	}

	val, err := strconv.Atoi(pathValue)
	if err != nil {
		return 0, fmt.Errorf(
			"path parameter %q must be an integer (got %q): %v: %w",
			key,
			pathValue,
			err,
			errs.ErrInvalidArgument,
		)
	}

	return val, nil
}
