package domain

import (
	"fmt"
	"strings"
	"unicode/utf8"

	errs "github.com/nseve/music-catalog/internal/core/errors"
)

type Artist struct {
	ID      int
	Name    string
	Country *string
}

func NewArtist(id int, name string, country *string) Artist {
	if country != nil {
		*country = strings.ToUpper(*country)
	}

	return Artist{
		ID:      id,
		Name:    name,
		Country: country,
	}
}

func NewArtistUninitialized(name string, country *string) Artist {
	return NewArtist(UninitializedID, name, country)
}

func (a *Artist) Validate() error {
	name := strings.TrimSpace(a.Name)
	nameLen := utf8.RuneCountInString(name)
	if nameLen < 1 || nameLen > 255 {
		return fmt.Errorf("invalid name length: %d: %w", nameLen, errs.ErrInvalidArgument)
	}

	if a.Country != nil {
		countryLen := len(*a.Country)
		if countryLen != 2 {
			return fmt.Errorf("invalid country code length: %d: must be exactly 2 characters long: %w", len(*a.Country), errs.ErrInvalidArgument)
		}
	}

	return nil
}

type ArtistPatch struct {
	Name    Nullable[string]
	Country Nullable[string]
}

func NewArtistPatch(name Nullable[string], country Nullable[string]) ArtistPatch {
	if country.Set && country.Value != nil {
		*country.Value = strings.ToUpper(*country.Value)
	}

	return ArtistPatch{
		Name:    name,
		Country: country,
	}
}

func (p *ArtistPatch) Validate() error {
	if p.Name.Set && p.Name.Value == nil {
		return fmt.Errorf("name cant be patched to NULL: %w", errs.ErrInvalidArgument)
	}

	return nil
}

func (a *Artist) ApplyPatch(patch ArtistPatch) error {
	if err := patch.Validate(); err != nil {
		return fmt.Errorf("validate artist patch: %w", err)
	}

	tmp := *a

	if patch.Name.Set {
		tmp.Name = *patch.Name.Value
	}

	if patch.Country.Set {
		tmp.Country = patch.Country.Value
	}

	if err := tmp.Validate(); err != nil {
		return fmt.Errorf("validate patched artist: %w", err)
	}

	*a = tmp

	return nil
}
