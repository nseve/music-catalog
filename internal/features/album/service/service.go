package service

import (
	"context"
	"fmt"

	"github.com/nseve/music-catalog/internal/core/domain"
)

type AlbumRepository interface {
	GetAll(ctx context.Context) ([]domain.Album, error)
	GetByID(ctx context.Context, id int) (domain.Album, error)
	GetByIDForUpdate(ctx context.Context, id int) (domain.Album, error)
	Create(ctx context.Context, album domain.Album) (domain.Album, error)
	Update(ctx context.Context, id int, album domain.Album) (domain.Album, error)
	Delete(ctx context.Context, id int) error
}

type TxManager interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type Service struct {
	albumRepository AlbumRepository
	txManager       TxManager
}

func New(albumRepository AlbumRepository, txManager TxManager) *Service {
	return &Service{
		albumRepository: albumRepository,
		txManager:       txManager,
	}
}

func (s *Service) GetAll(ctx context.Context) ([]domain.Album, error) {
	albums, err := s.albumRepository.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("get all albums: %w", err)
	}

	return albums, nil
}

func (s *Service) GetByID(ctx context.Context, id int) (domain.Album, error) {
	album, err := s.albumRepository.GetByID(ctx, id)
	if err != nil {
		return domain.Album{}, fmt.Errorf("get album by id: %w", err)
	}

	return album, nil
}

func (s *Service) Create(ctx context.Context, album domain.Album) (domain.Album, error) {
	if err := album.Validate(); err != nil {
		return domain.Album{}, fmt.Errorf("validate album: %w", err)
	}

	created, err := s.albumRepository.Create(ctx, album)
	if err != nil {
		return domain.Album{}, fmt.Errorf("create album: %w", err)
	}

	return created, nil
}

func (s *Service) Update(ctx context.Context, id int, patch domain.AlbumPatch) (domain.Album, error) {
	var updated domain.Album
	err := s.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		album, err := s.albumRepository.GetByIDForUpdate(ctx, id)
		if err != nil {
			return fmt.Errorf("get album by id: %w", err)
		}

		if err := album.ApplyPatch(patch); err != nil {
			return fmt.Errorf("apply album patch: %w", err)
		}

		updated, err = s.albumRepository.Update(ctx, id, album)
		if err != nil {
			return fmt.Errorf("update album: %w", err)
		}

		return nil
	})
	if err != nil {
		return domain.Album{}, err
	}

	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id int) error {
	if err := s.albumRepository.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete album: %w", err)
	}

	return nil
}
