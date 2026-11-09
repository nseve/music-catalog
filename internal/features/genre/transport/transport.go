package genre_transport

import (
	"context"
	"net/http"

	"github.com/nseve/music-catalog/internal/core/domain"
	"github.com/nseve/music-catalog/internal/core/logger"
	"github.com/nseve/music-catalog/internal/core/transport/request"
	"github.com/nseve/music-catalog/internal/core/transport/response"
	"github.com/nseve/music-catalog/internal/core/transport/server"
	"github.com/nseve/music-catalog/internal/core/utils"
)

type GenreService interface {
	GetAll(ctx context.Context) ([]domain.Genre, error)
	GetByID(ctx context.Context, id int) (domain.Genre, error)
	Create(ctx context.Context, genre domain.Genre) (domain.Genre, error)
	Update(ctx context.Context, genre domain.Genre) (domain.Genre, error)
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
	log := logger.FromContext(r.Context())
	responseHandler := response.NewResponseHandler(log, w)

	genreDomains, err := h.genreService.GetAll(r.Context())
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
	log := logger.FromContext(r.Context())
	responseHandler := response.NewResponseHandler(log, w)

	id, err := utils.GetIntPathValue(r, "id")
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to convert id to integer")
		return
	}

	genre, err := h.genreService.GetByID(r.Context(), id)
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
	log := logger.FromContext(r.Context())
	responseHandler := response.NewResponseHandler(log, w)

	var req GenreRequest
	if err := request.DecodeJSON(r, &req); err != nil {
		responseHandler.ErrorResponse(err, "failed to decode request")
		return
	}

	created, err := h.genreService.Create(r.Context(), req.ToDomain())
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
	log := logger.FromContext(r.Context())
	responseHandler := response.NewResponseHandler(log, w)

	id, err := utils.GetIntPathValue(r, "id")
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to convert id to integer")
		return
	}

	var req GenreRequest
	if err := request.DecodeJSON(r, &req); err != nil {
		responseHandler.ErrorResponse(err, "failed to decode request")
		return
	}

	genreDomain := req.ToDomain()
	genreDomain.ID = id

	updated, err := h.genreService.Update(r.Context(), genreDomain)
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
	log := logger.FromContext(r.Context())
	responseHandler := response.NewResponseHandler(log, w)

	id, err := utils.GetIntPathValue(r, "id")
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to convert id to integer")
		return
	}

	if err := h.genreService.Delete(r.Context(), id); err != nil {
		responseHandler.ErrorResponse(err, "failed to delete genre")
		return
	}

	responseHandler.NoContentResponse()
}

func (h *Handler) Routes() []server.Route {
	return []server.Route{
		{
			Method:  http.MethodGet,
			Path:    "/genre",
			Handler: h.GetAll,
		},
		{
			Method:  http.MethodGet,
			Path:    "/genre/{id}",
			Handler: h.GetByID,
		},
		{
			Method:  http.MethodPost,
			Path:    "/genre",
			Handler: h.Create,
		},
		{
			Method:  http.MethodPatch,
			Path:    "/genre/{id}",
			Handler: h.Update,
		},
		{
			Method:  http.MethodDelete,
			Path:    "/genre/{id}",
			Handler: h.Delete,
		},
	}
}
