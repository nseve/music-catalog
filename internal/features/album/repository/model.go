package repository

import (
	"time"

	"github.com/nseve/music-catalog/internal/core/domain"
)

type AlbumModel struct {
	ID          int
	ArtistID    int
	Title       string
	Type        string
	ReleaseDate time.Time
	CoverPath   *string
}

func (m *AlbumModel) ToDomain() domain.Album {
	return domain.NewAlbum(
		m.ID,
		m.ArtistID,
		m.Title,
		domain.AlbumType(m.Type),
		m.ReleaseDate,
		m.CoverPath,
	)
}
