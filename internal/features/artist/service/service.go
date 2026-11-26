package service

import (
	"context"
	"fmt"

	"github.com/nseve/music-catalog/internal/core/domain"
	errs "github.com/nseve/music-catalog/internal/core/errors"
)

type ArtistRepository interface {
	List(ctx context.Context, limit *int, offset *int) ([]domain.Artist, error)
	GetByID(ctx context.Context, id int) (domain.Artist, error)
	GetByIDForUpdate(ctx context.Context, id int) (domain.Artist, error)
	Create(ctx context.Context, artist domain.Artist) (domain.Artist, error)
	Update(ctx context.Context, id int, artist domain.Artist) (domain.Artist, error)
	Delete(ctx context.Context, id int) error
}

type TxManager interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type Service struct {
	artistRepository ArtistRepository
	txManager        TxManager
}

func New(artistRepository ArtistRepository, txManager TxManager) *Service {
	return &Service{
		artistRepository: artistRepository,
		txManager:        txManager,
	}
}

func (s *Service) List(ctx context.Context, limit *int, offset *int) ([]domain.Artist, error) {
	if limit != nil && *limit < 0 {
		return nil, fmt.Errorf("limit must be non-negative: %w", errs.ErrInvalidArgument)
	}

	if offset != nil && *offset < 0 {
		return nil, fmt.Errorf("offset must be non-negative: %w", errs.ErrInvalidArgument)
	}

	artists, err := s.artistRepository.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get all artists: %w", err)
	}

	return artists, nil
}

func (s *Service) GetByID(ctx context.Context, id int) (domain.Artist, error) {
	artist, err := s.artistRepository.GetByID(ctx, id)
	if err != nil {
		return domain.Artist{}, fmt.Errorf("get artist by id: %w", err)
	}

	return artist, nil
}

func (s *Service) Create(ctx context.Context, artist domain.Artist) (domain.Artist, error) {
	if err := artist.Validate(); err != nil {
		return domain.Artist{}, fmt.Errorf("validate artist: %w", err)
	}

	created, err := s.artistRepository.Create(ctx, artist)
	if err != nil {
		return domain.Artist{}, fmt.Errorf("create artist: %w", err)
	}

	return created, nil
}

func (s *Service) Update(ctx context.Context, id int, patch domain.ArtistPatch) (domain.Artist, error) {
	var updated domain.Artist

	err := s.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		artist, err := s.artistRepository.GetByIDForUpdate(ctx, id)
		if err != nil {
			return fmt.Errorf("get artist by id: %w", err)
		}

		if err := artist.ApplyPatch(patch); err != nil {
			return fmt.Errorf("apply artist patch: %w", err)
		}

		updated, err = s.artistRepository.Update(ctx, id, artist)
		if err != nil {
			return fmt.Errorf("update artist: %w", err)
		}

		return nil
	})
	if err != nil {
		return domain.Artist{}, err
	}

	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id int) error {
	if err := s.artistRepository.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete artist: %w", err)
	}

	return nil
}
