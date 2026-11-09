package genre_repository

import "github.com/nseve/music-catalog/internal/core/domain"

type GenreModel struct {
	ID   int
	Name string
}

func (m *GenreModel) ToDomain() domain.Genre {
	return domain.NewGenre(m.ID, m.Name)
}
