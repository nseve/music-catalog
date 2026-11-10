package genre_transport

import "github.com/nseve/music-catalog/internal/core/domain"

type GenreRequest struct {
	Name string `json:"name"`
}

func (r *GenreRequest) ToDomain() domain.Genre {
	return domain.NewGenreUninitialized(r.Name)
}

type GenreResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
