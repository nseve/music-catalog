package transport

import (
	"github.com/nseve/music-catalog/internal/core/domain"
	"github.com/nseve/music-catalog/internal/core/transport/types"
)

type CreateArtistRequest struct {
	Name    string  `json:"name"`
	Country *string `json:"country"`
}

func (r *CreateArtistRequest) ToDomain() domain.Artist {
	return domain.NewArtistUninitialized(r.Name, r.Country)
}

type PatchArtistRequest struct {
	Name    types.Nullable[string] `json:"name"`
	Country types.Nullable[string] `json:"country"`
}

func (p *PatchArtistRequest) ToDomain() domain.ArtistPatch {
	return domain.NewArtistPatch(
		p.Name.ToDomain(),
		p.Country.ToDomain(),
	)
}

type ArtistResponse struct {
	ID      int     `json:"id"`
	Name    string  `json:"name"`
	Country *string `json:"country"`
}
