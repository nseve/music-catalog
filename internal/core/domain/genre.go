package domain

import (
	"fmt"
	"strings"
	"unicode/utf8"

	errs "github.com/nseve/music-catalog/internal/core/errors"
)

type Genre struct {
	ID   int
	Name string
}

func NewGenre(id int, name string) Genre {
	return Genre{
		ID:   id,
		Name: name,
	}
}

func NewGenreUninitialized(name string) Genre {
	return NewGenre(UninitializedID, name)
}

func (g *Genre) Validate() error {
	name := strings.TrimSpace(g.Name)
	nameLen := utf8.RuneCountInString(name)
	if nameLen < 1 || nameLen > 255 {
		return fmt.Errorf("invalid name length: %d: %w", nameLen, errs.ErrInvalidArgument)
	}

	return nil
}

type GenrePatch struct {
	Name Nullable[string]
}

func NewGenrePatch(name Nullable[string]) GenrePatch {
	return GenrePatch{
		Name: name,
	}
}

func (p *GenrePatch) Validate() error {
	if p.Name.Set && p.Name.Value == nil {
		return fmt.Errorf("name cant be patched to NULL: %w", errs.ErrInvalidArgument)
	}

	return nil
}

func (g *Genre) ApplyPatch(patch GenrePatch) error {
	if err := patch.Validate(); err != nil {
		return fmt.Errorf("validate genre patch: %w", err)
	}

	tmp := *g

	if patch.Name.Set {
		tmp.Name = *patch.Name.Value
	}

	if err := tmp.Validate(); err != nil {
		return fmt.Errorf("validate patched genre: %w", err)
	}

	*g = tmp

	return nil
}
