package request

import (
	"fmt"
	"mime/multipart"
	"net/http"

	errs "github.com/nseve/music-catalog/internal/core/errors"
)

const (
	multipartOverhead = 1 << 20
	multipartMemory   = 1 << 20
)

func ParseFile(w http.ResponseWriter, r *http.Request, field string, maxSize int64) (multipart.File, error) {
	r.Body = http.MaxBytesReader(w, r.Body, maxSize+multipartOverhead)
	if err := r.ParseMultipartForm(multipartMemory); err != nil {
		return nil, fmt.Errorf("parse multipart form: %v: %w", err, errs.ErrInvalidArgument)
	}

	file, header, err := r.FormFile(field)
	if err != nil {
		return nil, fmt.Errorf("form file %q: %v: %w", field, err, errs.ErrInvalidArgument)
	}

	if header.Size > maxSize {
		_ = file.Close()
		return nil, fmt.Errorf("file too large: %d bytes, max %d: %w", header.Size, maxSize, errs.ErrInvalidArgument)
	}

	return file, nil
}
