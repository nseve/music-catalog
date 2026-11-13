package transport

import (
	"context"
	"net/http"

	"github.com/nseve/music-catalog/internal/core/domain"
	"github.com/nseve/music-catalog/internal/core/logger"
	"github.com/nseve/music-catalog/internal/core/transport/request"
	"github.com/nseve/music-catalog/internal/core/transport/response"
	"github.com/nseve/music-catalog/internal/core/transport/server"
	"github.com/nseve/music-catalog/internal/core/transport/utils"
)

type GenreService interface {
	GetAll(ctx context.Context) ([]domain.Genre, error)
	GetByID(ctx context.Context, id int) (domain.Genre, error)
	Create(ctx context.Context, genre domain.Genre) (domain.Genre, error)
	Update(ctx context.Context, id int, patch domain.GenrePatch) (domain.Genre, error)
	Delete(ctx context.Context, id int) error
}

type Handler struct {
	genreService GenreService
}

func New(genreService GenreService) *Handler {
	return &Handler{
		genreService: genreService,
	}
}

func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := response.NewResponseHandler(log, w)

	genreDomains, err := h.genreService.GetAll(ctx)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get all genres")
		return
	}

	genreResponses := make([]GenreResponse, 0, len(genreDomains))
	for _, d := range genreDomains {
		genreResponses = append(genreResponses, GenreResponse{
			ID:   d.ID,
			Name: d.Name,
		})
	}

	responseHandler.JSONResponse(genreResponses, http.StatusOK)
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

	genre, err := h.genreService.GetByID(ctx, id)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get genre by id")
		return
	}

	responseHandler.JSONResponse(GenreResponse{
		ID:   genre.ID,
		Name: genre.Name,
	}, http.StatusOK)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := response.NewResponseHandler(log, w)

	var req CreateGenreRequest
	if err := request.DecodeJSON(r, &req); err != nil {
		responseHandler.ErrorResponse(err, "failed to decode request")
		return
	}

	created, err := h.genreService.Create(ctx, req.ToDomain())
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to create genre")
		return
	}

	responseHandler.JSONResponse(GenreResponse{
		ID:   created.ID,
		Name: created.Name,
	}, http.StatusCreated)
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

	var req PatchGenreRequest
	if err := request.DecodeJSON(r, &req); err != nil {
		responseHandler.ErrorResponse(err, "failed to decode request")
		return
	}

	updated, err := h.genreService.Update(ctx, id, req.ToDomain())
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to update genre")
		return
	}

	responseHandler.JSONResponse(GenreResponse{
		ID:   updated.ID,
		Name: updated.Name,
	}, http.StatusOK)
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

	if err := h.genreService.Delete(ctx, id); err != nil {
		responseHandler.ErrorResponse(err, "failed to delete genre")
		return
	}

	responseHandler.NoContentResponse()
}

func (h *Handler) Routes() []server.Route {
	return []server.Route{
		{
			Method:  http.MethodGet,
			Path:    "/genres",
			Handler: h.GetAll,
		},
		{
			Method:  http.MethodGet,
			Path:    "/genres/{id}",
			Handler: h.GetByID,
		},
		{
			Method:  http.MethodPost,
			Path:    "/genres",
			Handler: h.Create,
		},
		{
			Method:  http.MethodPatch,
			Path:    "/genres/{id}",
			Handler: h.Update,
		},
		{
			Method:  http.MethodDelete,
			Path:    "/genres/{id}",
			Handler: h.Delete,
		},
	}
}
