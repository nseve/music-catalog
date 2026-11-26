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

type ArtistService interface {
	List(ctx context.Context, limit *int, offset *int) ([]domain.Artist, error)
	GetByID(ctx context.Context, id int) (domain.Artist, error)
	Create(ctx context.Context, artist domain.Artist) (domain.Artist, error)
	Update(ctx context.Context, id int, patch domain.ArtistPatch) (domain.Artist, error)
	Delete(ctx context.Context, id int) error
}

type Handler struct {
	artistService ArtistService
}

func New(artistService ArtistService) *Handler {
	return &Handler{
		artistService: artistService,
	}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := response.NewResponseHandler(log, w)

	limit, offset, err := getLimitOffsetQueryParams(r)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get query parameters")
		return
	}

	artistDomains, err := h.artistService.List(ctx, limit, offset)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to list artists")
		return
	}

	artistResponses := make([]ArtistResponse, 0, len(artistDomains))
	for _, d := range artistDomains {
		artistResponses = append(artistResponses, ArtistResponse{
			ID:      d.ID,
			Name:    d.Name,
			Country: d.Country,
		})
	}

	responseHandler.JSONResponse(artistResponses, http.StatusOK)
}

func getLimitOffsetQueryParams(r *http.Request) (*int, *int, error) {
	const (
		limitQueryParamKey  = "limit"
		offsetQueryParamKey = "offset"
	)

	limit, err := utils.GetIntQueryParam(r, limitQueryParamKey)
	if err != nil {
		return nil, nil, err
	}

	offset, err := utils.GetIntQueryParam(r, offsetQueryParamKey)
	if err != nil {
		return nil, nil, err
	}

	return limit, offset, nil
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := response.NewResponseHandler(log, w)

	id, err := utils.GetIntPathValue(r, "id")
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get path parameter \"id\"")
		return
	}

	artist, err := h.artistService.GetByID(ctx, id)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get artist")
		return
	}

	responseHandler.JSONResponse(ArtistResponse{
		ID:      artist.ID,
		Name:    artist.Name,
		Country: artist.Country,
	}, http.StatusOK)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := response.NewResponseHandler(log, w)

	var req CreateArtistRequest
	if err := request.DecodeJSON(r, &req); err != nil {
		responseHandler.ErrorResponse(err, "failed to decode request body")
		return
	}

	created, err := h.artistService.Create(ctx, req.ToDomain())
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to create artist")
		return
	}

	responseHandler.JSONResponse(ArtistResponse{
		ID:      created.ID,
		Name:    created.Name,
		Country: created.Country,
	}, http.StatusCreated)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := response.NewResponseHandler(log, w)

	id, err := utils.GetIntPathValue(r, "id")
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get path parameter \"id\"")
		return
	}

	var req PatchArtistRequest
	if err := request.DecodeJSON(r, &req); err != nil {
		responseHandler.ErrorResponse(err, "failed to decode request body")
		return
	}

	updated, err := h.artistService.Update(ctx, id, req.ToDomain())
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to update artist")
		return
	}

	responseHandler.JSONResponse(ArtistResponse{
		ID:      updated.ID,
		Name:    updated.Name,
		Country: updated.Country,
	}, http.StatusOK)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := response.NewResponseHandler(log, w)

	id, err := utils.GetIntPathValue(r, "id")
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get path parameter \"id\"")
		return
	}

	if err := h.artistService.Delete(ctx, id); err != nil {
		responseHandler.ErrorResponse(err, "failed to delete artist")
		return
	}

	responseHandler.NoContentResponse()
}

func (h *Handler) Routes() []server.Route {
	return []server.Route{
		{
			Method:  http.MethodGet,
			Path:    "/artists",
			Handler: h.List,
		},
		{
			Method:  http.MethodGet,
			Path:    "/artists/{id}",
			Handler: h.GetByID,
		},
		{
			Method:  http.MethodPost,
			Path:    "/artists",
			Handler: h.Create,
		},
		{
			Method:  http.MethodPatch,
			Path:    "/artists/{id}",
			Handler: h.Update,
		},
		{
			Method:  http.MethodDelete,
			Path:    "/artists/{id}",
			Handler: h.Delete,
		},
	}
}
