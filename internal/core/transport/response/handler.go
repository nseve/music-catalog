package response

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
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
	response := map[string]string{
		"message": msg,
		"error":   err.Error(),
	}

	h.JSONResponse(response, http.StatusInternalServerError)
}

func (h *ResponseHandler) JSONResponse(body any, statusCode int) {
	h.w.Header().Set("Content-Type", "application/json")
	h.w.WriteHeader(statusCode)

	if err := json.NewEncoder(h.w).Encode(body); err != nil {
		h.log.Error("encode JSON response", slog.Any("error", err))
	}
}
