package transport

import (
	"context"
	"io"
	"net/http"
	"os"

	"github.com/nseve/music-catalog/internal/core/domain"
	"github.com/nseve/music-catalog/internal/core/logger"
	"github.com/nseve/music-catalog/internal/core/transport/request"
	"github.com/nseve/music-catalog/internal/core/transport/response"
	"github.com/nseve/music-catalog/internal/core/transport/server"
	"github.com/nseve/music-catalog/internal/core/transport/utils"
)

type AlbumService interface {
	List(ctx context.Context, artistID *int, limit *int, offset *int) ([]domain.Album, error)
	Create(ctx context.Context, album domain.Album) (domain.Album, error)
	GetByID(ctx context.Context, id int) (domain.Album, error)
	Update(ctx context.Context, id int, patch domain.AlbumPatch) (domain.Album, error)
	Delete(ctx context.Context, id int) error

	SetCover(ctx context.Context, id int, src io.Reader) (domain.Album, error)
	GetCover(ctx context.Context, id int) (*os.File, error)
	DeleteCover(ctx context.Context, id int) error
}

type Handler struct {
	albumService AlbumService
}

func New(albumService AlbumService) *Handler {
	return &Handler{
		albumService: albumService,
	}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := response.NewResponseHandler(log, w)

	queryParams, err := getArtistIDLimitOffsetQueryParams(r)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get query parameters")
		return
	}

	albumDomains, err := h.albumService.List(ctx, queryParams.artistID, queryParams.limit, queryParams.offset)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to list albums")
		return
	}

	albumResponses := make([]AlbumResponse, 0, len(albumDomains))
	for _, d := range albumDomains {
		albumResponses = append(albumResponses, ResponseFromDomain(d))
	}

	responseHandler.JSONResponse(albumResponses, http.StatusOK)
}

type queryParams struct {
	artistID *int
	limit    *int
	offset   *int
}

func getArtistIDLimitOffsetQueryParams(r *http.Request) (queryParams, error) {
	const (
		artistIDQueryParamKey = "artist_id"
		limitQueryParamKey    = "limit"
		offsetQueryParamKey   = "offset"
	)

	artistID, err := utils.GetIntQueryParam(r, artistIDQueryParamKey)
	if err != nil {
		return queryParams{}, err
	}

	limit, err := utils.GetIntQueryParam(r, limitQueryParamKey)
	if err != nil {
		return queryParams{}, err
	}

	offset, err := utils.GetIntQueryParam(r, offsetQueryParamKey)
	if err != nil {
		return queryParams{}, err
	}

	return queryParams{
		artistID: artistID,
		limit:    limit,
		offset:   offset,
	}, nil
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

	albumDomain, err := h.albumService.GetByID(ctx, id)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get album")
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
		responseHandler.ErrorResponse(err, "failed to decode request body")
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
		responseHandler.ErrorResponse(err, "failed to get path parameter \"id\"")
		return
	}

	var req PatchAlbumRequest
	if err := request.DecodeJSON(r, &req); err != nil {
		responseHandler.ErrorResponse(err, "failed to decode request body")
		return
	}

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
		responseHandler.ErrorResponse(err, "failed to get path parameter \"id\"")
		return
	}

	if err := h.albumService.Delete(ctx, id); err != nil {
		responseHandler.ErrorResponse(err, "failed to delete album")
		return
	}

	responseHandler.NoContentResponse()
}

func (h *Handler) SetCover(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := response.NewResponseHandler(log, w)

	id, err := utils.GetIntPathValue(r, "id")
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get path parameter \"id\"")
		return
	}

	file, err := request.ParseFile(w, r, "cover", domain.MaxCoverSize)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to read cover")
		return
	}
	defer file.Close()

	updated, err := h.albumService.SetCover(ctx, id, file)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to set album cover")
		return
	}

	responseHandler.JSONResponse(ResponseFromDomain(updated), http.StatusOK)
}

func (h *Handler) GetCover(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := response.NewResponseHandler(log, w)

	id, err := utils.GetIntPathValue(r, "id")
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get path parameter \"id\"")
		return
	}

	f, err := h.albumService.GetCover(ctx, id)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get album cover")
		return
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to stat cover file")
		return
	}

	w.Header().Set("Cache-Control", "no-cache")
	http.ServeContent(w, r, f.Name(), info.ModTime(), f)
}

func (h *Handler) DeleteCover(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := response.NewResponseHandler(log, w)

	id, err := utils.GetIntPathValue(r, "id")
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get path parameter \"id\"")
		return
	}

	if err := h.albumService.DeleteCover(ctx, id); err != nil {
		responseHandler.ErrorResponse(err, "failed to delete album cover")
		return
	}

	responseHandler.NoContentResponse()
}

func (h *Handler) Routes() []server.Route {
	return []server.Route{
		{
			Method:  http.MethodGet,
			Path:    "/albums",
			Handler: h.List,
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
		{
			Method:  http.MethodPut,
			Path:    "/albums/{id}/cover",
			Handler: h.SetCover,
		},
		{
			Method:  http.MethodGet,
			Path:    "/albums/{id}/cover",
			Handler: h.GetCover,
		},
		{
			Method:  http.MethodDelete,
			Path:    "/albums/{id}/cover",
			Handler: h.DeleteCover,
		},
	}
}
