package transport

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/nseve/music-catalog/internal/core/domain"
	"github.com/nseve/music-catalog/internal/core/logger"
	"github.com/nseve/music-catalog/internal/core/transport/request"
	"github.com/nseve/music-catalog/internal/core/transport/response"
	"github.com/nseve/music-catalog/internal/core/transport/server"
	"github.com/nseve/music-catalog/internal/core/transport/utils"
)

type AlbumService interface {
	GetAll(ctx context.Context) ([]domain.Album, error)
	Create(ctx context.Context, album domain.Album) (domain.Album, error)
	GetByID(ctx context.Context, id int) (domain.Album, error)
	Update(ctx context.Context, id int, patch domain.AlbumPatch) (domain.Album, error)
	Delete(ctx context.Context, id int) error
}

type Handler struct {
	albumService AlbumService
}

func New(albumService AlbumService) *Handler {
	return &Handler{
		albumService: albumService,
	}
}

func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := response.NewResponseHandler(log, w)

	albumDomains, err := h.albumService.GetAll(ctx)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get all albums")
		return
	}

	albumResponses := make([]AlbumResponse, 0, len(albumDomains))
	for _, d := range albumDomains {
		albumResponses = append(albumResponses, ResponseFromDomain(d))
	}

	responseHandler.JSONResponse(albumResponses, http.StatusOK)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := response.NewResponseHandler(log, w)

	id, err := utils.GetIntPathValue(r, "id")
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get path param value")
		return
	}

	albumDomain, err := h.albumService.GetByID(ctx, id)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get album by id")
		return
	}

	responseHandler.JSONResponse(ResponseFromDomain(albumDomain), http.StatusOK)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := response.NewResponseHandler(log, w)

	var req CreateAlbumRequest
	if err := request.DecodeJSON(r, &req); err != nil {
		responseHandler.ErrorResponse(err, "failed to decode request")
		return
	}

	created, err := h.albumService.Create(ctx, req.ToDomain())
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to create album")
		return
	}

	responseHandler.JSONResponse(ResponseFromDomain(created), http.StatusCreated)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := response.NewResponseHandler(log, w)

	id, err := utils.GetIntPathValue(r, "id")
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get path param value")
		return
	}

	var req PatchAlbumRequest
	if err := request.DecodeJSON(r, &req); err != nil {
		responseHandler.ErrorResponse(err, "failed to decode request")
		return
	}

	log.Debug("got request", slog.Any("req", req))

	updated, err := h.albumService.Update(ctx, id, req.ToDomain())
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to update album")
		return
	}

	responseHandler.JSONResponse(ResponseFromDomain(updated), http.StatusOK)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := response.NewResponseHandler(log, w)

	id, err := utils.GetIntPathValue(r, "id")
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get path param value")
		return
	}

	if err := h.albumService.Delete(ctx, id); err != nil {
		responseHandler.ErrorResponse(err, "failed to delete album")
		return
	}

	responseHandler.NoContentResponse()
}

func (h *Handler) Routes() []server.Route {
	return []server.Route{
		{
			Method:  http.MethodGet,
			Path:    "/albums",
			Handler: h.GetAll,
		},
		{
			Method:  http.MethodGet,
			Path:    "/albums/{id}",
			Handler: h.GetByID,
		},
		{
			Method:  http.MethodPost,
			Path:    "/albums",
			Handler: h.Create,
		},
		{
			Method:  http.MethodPatch,
			Path:    "/albums/{id}",
			Handler: h.Update,
		},
		{
			Method:  http.MethodDelete,
			Path:    "/albums/{id}",
			Handler: h.Delete,
		},
	}
}
