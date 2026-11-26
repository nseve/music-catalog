package service

import (
	"context"
	"fmt"

	"github.com/nseve/music-catalog/internal/core/domain"
	errs "github.com/nseve/music-catalog/internal/core/errors"
)

type GenreRepository interface {
	List(ctx context.Context, limit *int, offset *int) ([]domain.Genre, error)
	GetByID(ctx context.Context, id int) (domain.Genre, error)
	GetByIDForUpdate(ctx context.Context, id int) (domain.Genre, error)
	Create(ctx context.Context, genre domain.Genre) (domain.Genre, error)
	Update(ctx context.Context, id int, genre domain.Genre) (domain.Genre, error)
	Delete(ctx context.Context, id int) error
}

type TxManager interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type Service struct {
	genreRepository GenreRepository
	txManager       TxManager
}

func New(genreRepository GenreRepository, txManager TxManager) *Service {
	return &Service{
		genreRepository: genreRepository,
		txManager:       txManager,
	}
}

func (s *Service) List(ctx context.Context, limit *int, offset *int) ([]domain.Genre, error) {
	if limit != nil && *limit < 0 {
		return nil, fmt.Errorf("limit must be non-negative: %w", errs.ErrInvalidArgument)
	}

	if offset != nil && *offset < 0 {
		return nil, fmt.Errorf("offset must be non-negative: %w", errs.ErrInvalidArgument)
	}

	genres, err := s.genreRepository.List(ctx, limit, offset)
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

func (s *Service) Update(ctx context.Context, id int, patch domain.GenrePatch) (domain.Genre, error) {
	var updated domain.Genre

	err := s.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		genre, err := s.genreRepository.GetByIDForUpdate(ctx, id)
		if err != nil {
			return fmt.Errorf("get genre by id: %w", err)
		}

		if err := genre.ApplyPatch(patch); err != nil {
			return fmt.Errorf("apply genre patch: %w", err)
		}

		updated, err = s.genreRepository.Update(ctx, id, genre)
		if err != nil {
			return fmt.Errorf("update genre: %w", err)
		}

		return nil
	})
	if err != nil {
		return domain.Genre{}, err
	}

	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id int) error {
	if err := s.genreRepository.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete genre: %w", err)
	}

	return nil
}
