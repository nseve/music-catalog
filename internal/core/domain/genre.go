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
		return fmt.Errorf("invalid Name length: %d: %w", nameLen, errs.ErrInvalidArgument)
	}

	return nil
}
