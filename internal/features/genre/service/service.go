package genre_service

import (
	"context"
	"fmt"

	"github.com/nseve/music-catalog/internal/core/domain"
)

type GenreRepository interface {
	GetAll(ctx context.Context) ([]domain.Genre, error)
	GetByID(ctx context.Context, id int) (domain.Genre, error)
	Create(ctx context.Context, genre domain.Genre) (domain.Genre, error)
	Update(ctx context.Context, genre domain.Genre) (domain.Genre, error)
	Delete(ctx context.Context, id int) error
}

type Service struct {
	genreRepository GenreRepository
}

func New(genreRepository GenreRepository) *Service {
	return &Service{
		genreRepository: genreRepository,
	}
}

func (s *Service) GetAll(ctx context.Context) ([]domain.Genre, error) {
	genres, err := s.genreRepository.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("get all genres: %w", err)
	}

	return genres, nil
}

func (s *Service) GetByID(ctx context.Context, id int) (domain.Genre, error) {
	genre, err := s.genreRepository.GetByID(ctx, id)
	if err != nil {
		return domain.Genre{}, fmt.Errorf("get genre by id: %w", err)
	}

	return genre, nil
}

func (s *Service) Create(ctx context.Context, genre domain.Genre) (domain.Genre, error) {
	if err := genre.Validate(); err != nil {
		return domain.Genre{}, fmt.Errorf("validate genre: %w", err)
	}

	created, err := s.genreRepository.Create(ctx, genre)
	if err != nil {
		return domain.Genre{}, fmt.Errorf("create genre: %w", err)
	}

	return created, nil
}

func (s *Service) Update(ctx context.Context, genre domain.Genre) (domain.Genre, error) {
	if err := genre.Validate(); err != nil {
		return domain.Genre{}, fmt.Errorf("validate genre: %w", err)
	}

	updated, err := s.genreRepository.Update(ctx, genre)
	if err != nil {
		return domain.Genre{}, fmt.Errorf("update genre: %w", err)
	}

	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id int) error {
	if err := s.genreRepository.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete genre: %w", err)
	}

	return nil
}
