package domain

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	errs "github.com/nseve/music-catalog/internal/core/errors"
)

type AlbumType string

const (
	AlbumTypeAlbum       AlbumType = "album"
	AlbumTypeSingle      AlbumType = "single"
	AlbumTypeEp          AlbumType = "ep"
	AlbumTypeCompilation AlbumType = "compilation"
)

func (t AlbumType) IsValid() bool {
	switch t {
	case AlbumTypeAlbum, AlbumTypeSingle, AlbumTypeEp, AlbumTypeCompilation:
		return true
	}
	return false
}

type Album struct {
	ID          int
	ArtistID    int
	Title       string
	Type        AlbumType
	ReleaseDate time.Time
	CoverKey    *string
}

func NewAlbum(
	id int,
	artistID int,
	title string,
	albumType AlbumType,
	releaseDate time.Time,
	coverKey *string,
) Album {
	albumType = AlbumType(strings.ToLower(string(albumType)))
	return Album{
		ID:          id,
		ArtistID:    artistID,
		Title:       title,
		Type:        albumType,
		ReleaseDate: releaseDate,
		CoverKey:    coverKey,
	}
}

func NewAlbumUninitialized(
	artistID int,
	title string,
	albumType AlbumType,
	releaseDate time.Time,
) Album {
	return NewAlbum(
		UninitializedID,
		artistID,
		title,
		albumType,
		releaseDate,
		nil,
	)
}

func (a *Album) Validate() error {
	title := strings.TrimSpace(a.Title)
	titleLen := utf8.RuneCountInString(title)
	if titleLen < 1 || titleLen > 255 {
		return fmt.Errorf("invalid title length: %d: %w", titleLen, errs.ErrInvalidArgument)
	}

	if !a.Type.IsValid() {
		return fmt.Errorf("invalid album type: %s: %w", a.Type, errs.ErrInvalidArgument)
	}

	return nil
}

type AlbumPatch struct {
	ArtistID    Nullable[int]
	Title       Nullable[string]
	Type        Nullable[AlbumType]
	ReleaseDate Nullable[time.Time]
}

func NewAlbumPatch(
	artistID Nullable[int],
	title Nullable[string],
	albumType Nullable[AlbumType],
	releaseDate Nullable[time.Time],
) AlbumPatch {
	if albumType.Set && albumType.Value != nil {
		*albumType.Value = AlbumType(strings.ToLower(string(*albumType.Value)))
	}

	return AlbumPatch{
		ArtistID:    artistID,
		Title:       title,
		Type:        albumType,
		ReleaseDate: releaseDate,
	}
}

func (p *AlbumPatch) Validate() error {
	if p.ArtistID.Set && p.ArtistID.Value == nil {
		return fmt.Errorf("artist_id cant be patched to NULL: %w", errs.ErrInvalidArgument)
	}

	if p.Title.Set && p.Title.Value == nil {
		return fmt.Errorf("title cant be patched to NULL: %w", errs.ErrInvalidArgument)
	}

	if p.Type.Set && p.Type.Value == nil {
		return fmt.Errorf("album_type cant be patched to NULL: %w", errs.ErrInvalidArgument)
	}

	if p.ReleaseDate.Set && p.ReleaseDate.Value == nil {
		return fmt.Errorf("release_date cant be patched to NULL: %w", errs.ErrInvalidArgument)
	}

	return nil
}

func (a *Album) ApplyPatch(patch AlbumPatch) error {
	if err := patch.Validate(); err != nil {
		return fmt.Errorf("validate album patch: %w", err)
	}

	tmp := *a

	if patch.ArtistID.Set {
		tmp.ArtistID = *patch.ArtistID.Value
	}

	if patch.Title.Set {
		tmp.Title = *patch.Title.Value
	}

	if patch.Type.Set {
		tmp.Type = *patch.Type.Value
	}

	if patch.ReleaseDate.Set {
		tmp.ReleaseDate = *patch.ReleaseDate.Value
	}

	if err := tmp.Validate(); err != nil {
		return fmt.Errorf("validate patched album: %w", err)
	}

	*a = tmp

	return nil
}
