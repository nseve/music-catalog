package utils

import (
	"fmt"
	"net/http"
	"strconv"

	errs "github.com/nseve/music-catalog/internal/core/errors"
)

func GetIntPathValue(r *http.Request, key string) (int, error) {
	val, err := strconv.Atoi(r.PathValue(key))
	if err != nil {
		return 0, fmt.Errorf("failed to convert path value to integer: %v: %w", err, errs.ErrInvalidArgument)
	}

	return val, nil
}
