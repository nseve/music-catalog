package request

import (
	"encoding/json"
	"fmt"
	"net/http"

	errs "github.com/nseve/music-catalog/internal/core/errors"
)

func DecodeJSON(r *http.Request, dest any) error {
	if err := json.NewDecoder(r.Body).Decode(dest); err != nil {
		return fmt.Errorf("decode json: %v: %w", err, errs.ErrInvalidArgument)
	}

	return nil
}
