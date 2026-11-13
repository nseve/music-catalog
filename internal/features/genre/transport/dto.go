package transport

import (
	"github.com/nseve/music-catalog/internal/core/domain"
	"github.com/nseve/music-catalog/internal/core/transport/types"
)

type CreateGenreRequest struct {
	Name string `json:"name"`
}

func (r *CreateGenreRequest) ToDomain() domain.Genre {
	return domain.NewGenreUninitialized(r.Name)
}

type PatchGenreRequest struct {
	Name types.Nullable[string] `json:"name"`
}

func (p *PatchGenreRequest) ToDomain() domain.GenrePatch {
	return domain.NewGenrePatch(p.Name.ToDomain())
}

type GenreResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
