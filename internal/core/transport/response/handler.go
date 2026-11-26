package response

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	errs "github.com/nseve/music-catalog/internal/core/errors"
)

type ResponseHandler struct {
	log *slog.Logger
	w   http.ResponseWriter
}

func NewResponseHandler(log *slog.Logger, w http.ResponseWriter) *ResponseHandler {
	return &ResponseHandler{
		log: log,
		w:   w,
	}
}

func (h *ResponseHandler) PanicResponse(p any, msg string) {
	err := fmt.Errorf("unexpected panic: %v", p)

	h.log.Error(msg, slog.Any("panic", err))

	body := ErrorResponse{
		Message: msg,
		Error:   err.Error(),
	}

	h.JSONResponse(body, http.StatusInternalServerError)
}

func (h *ResponseHandler) JSONResponse(body any, statusCode int) {
	h.w.Header().Set("Content-Type", "application/json")
	h.w.WriteHeader(statusCode)

	if err := json.NewEncoder(h.w).Encode(body); err != nil {
		h.log.Error("encode JSON response", slog.Any("error", err))
	}
}

func (h *ResponseHandler) ErrorResponse(err error, msg string) {
	var (
		statusCode int
		logFunc    func(string, ...any)
	)

	switch {
	case errors.Is(err, errs.ErrInvalidArgument):
		statusCode = http.StatusBadRequest
		logFunc = h.log.Warn
	case errors.Is(err, errs.ErrNotFound):
		statusCode = http.StatusNotFound
		logFunc = h.log.Debug
	case errors.Is(err, errs.ErrConflict):
		statusCode = http.StatusConflict
		logFunc = h.log.Debug
	case errors.Is(err, errs.ErrUnprocessableEntity):
		statusCode = http.StatusUnprocessableEntity
		logFunc = h.log.Debug
	default:
		statusCode = http.StatusInternalServerError
		logFunc = h.log.Error
	}

	logFunc(msg, slog.Any("error", err))

	body := ErrorResponse{
		Message: msg,
		Error:   err.Error(),
	}

	h.JSONResponse(body, statusCode)
}

func (h *ResponseHandler) NoContentResponse() {
	h.w.WriteHeader(http.StatusNoContent)
}
