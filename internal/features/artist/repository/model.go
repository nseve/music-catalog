package repository

import "github.com/nseve/music-catalog/internal/core/domain"

type ArtistModel struct {
	ID      int
	Name    string
	Country *string
}

func (m *ArtistModel) ToDomain() domain.Artist {
	return domain.NewArtist(m.ID, m.Name, m.Country)
}
