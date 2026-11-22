package transport

import (
	"time"

	"github.com/nseve/music-catalog/internal/core/domain"
	"github.com/nseve/music-catalog/internal/core/transport/types"
)

type CreateAlbumRequest struct {
	ArtistID    int        `json:"artist_id"`
	Title       string     `json:"title"`
	Type        string     `json:"type"`
	ReleaseDate types.Date `json:"release_date"`
}

func (r *CreateAlbumRequest) ToDomain() domain.Album {
	return domain.NewAlbumUninitialized(
		r.ArtistID,
		r.Title,
		domain.AlbumType(r.Type),
		r.ReleaseDate.Time,
	)
}

type PatchAlbumRequest struct {
	ArtistID    types.Nullable[int]        `json:"artist_id"`
	Title       types.Nullable[string]     `json:"title"`
	Type        types.Nullable[string]     `json:"type"`
	ReleaseDate types.Nullable[types.Date] `json:"release_date"`
}

func (p *PatchAlbumRequest) ToDomain() domain.AlbumPatch {
	albumType := types.MapNullable(
		p.Type.ToDomain(),
		func(s string) domain.AlbumType {
			return domain.AlbumType(s)
		},
	)

	releaseDate := types.MapNullable(
		p.ReleaseDate.ToDomain(),
		func(d types.Date) time.Time {
			return d.Time
		},
	)

	return domain.NewAlbumPatch(
		p.ArtistID.ToDomain(),
		p.Title.ToDomain(),
		albumType,
		releaseDate,
	)
}

type AlbumResponse struct {
	ID          int        `json:"id"`
	ArtistID    int        `json:"artist_id"`
	Title       string     `json:"title"`
	Type        string     `json:"type"`
	ReleaseDate types.Date `json:"release_date"`
}

func ResponseFromDomain(album domain.Album) AlbumResponse {
	return AlbumResponse{
		ID:          album.ID,
		ArtistID:    album.ArtistID,
		Title:       album.Title,
		Type:        string(album.Type),
		ReleaseDate: types.Date{Time: album.ReleaseDate},
	}
}
