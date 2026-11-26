package domain

import (
	"fmt"

	errs "github.com/nseve/music-catalog/internal/core/errors"
)

const MaxCoverSize = 5 << 20

func CoverExt(contentType string) (string, error) {
	switch contentType {
	case "image/jpeg":
		return ".jpg", nil
	case "image/png":
		return ".png", nil
	case "image/webp":
		return ".webp", nil
	}

	return "", fmt.Errorf("unsupported cover type %q: %w", contentType, errs.ErrInvalidArgument)
}
