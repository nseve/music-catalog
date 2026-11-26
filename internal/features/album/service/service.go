package service

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/nseve/music-catalog/internal/core/domain"
	errs "github.com/nseve/music-catalog/internal/core/errors"
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

type CoverStorage interface {
	Save(key string, src io.Reader) error
	Open(key string) (*os.File, error)
	Discard(key string)
}

type Service struct {
	albumRepository AlbumRepository
	txManager       TxManager
	coverStorage    CoverStorage
}

func New(albumRepository AlbumRepository, txManager TxManager, coverStorage CoverStorage) *Service {
	return &Service{
		albumRepository: albumRepository,
		txManager:       txManager,
		coverStorage:    coverStorage,
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
	album, err := s.albumRepository.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get album by id: %w", err)
	}

	if err := s.albumRepository.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete album: %w", err)
	}

	if album.CoverKey != nil {
		s.coverStorage.Discard(*album.CoverKey)
	}

	return nil
}

func (s *Service) SetCover(ctx context.Context, id int, src io.Reader) (domain.Album, error) {
	br := bufio.NewReader(src)
	head, err := br.Peek(512)
	if err != nil && !errors.Is(err, io.EOF) {
		return domain.Album{}, fmt.Errorf("read cover: %w", err)
	}

	if len(head) == 0 {
		return domain.Album{}, fmt.Errorf("empty cover file: %w", errs.ErrInvalidArgument)
	}

	ext, err := domain.CoverExt(http.DetectContentType(head))
	if err != nil {
		return domain.Album{}, err
	}

	key := uuid.NewString() + ext
	if err := s.coverStorage.Save(key, br); err != nil {
		return domain.Album{}, fmt.Errorf("save cover: %w", err)
	}

	var updated domain.Album
	var oldCover *string
	err = s.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		album, err := s.albumRepository.GetByIDForUpdate(ctx, id)
		if err != nil {
			return fmt.Errorf("get album by id: %w", err)
		}
		oldCover = album.CoverKey

		album.CoverKey = &key
		fmt.Println("album before update:", album)
		updated, err = s.albumRepository.Update(ctx, id, album)
		if err != nil {
			return fmt.Errorf("update cover key: %w", err)
		}

		return nil
	})
	if err != nil {
		s.coverStorage.Discard(key)
		return domain.Album{}, err
	}

	if oldCover != nil {
		s.coverStorage.Discard(*oldCover)
	}

	return updated, nil
}

func (s *Service) GetCover(ctx context.Context, id int) (*os.File, error) {
	album, err := s.albumRepository.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get album by id: %w", err)
	}

	if album.CoverKey == nil {
		return nil, fmt.Errorf("album %d has no cover: %w", id, errs.ErrNotFound)
	}

	return s.coverStorage.Open(*album.CoverKey)
}

func (s *Service) DeleteCover(ctx context.Context, id int) error {
	var oldCover *string
	err := s.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		album, err := s.albumRepository.GetByIDForUpdate(ctx, id)
		if err != nil {
			return fmt.Errorf("get album by id: %w", err)
		}

		if album.CoverKey == nil {
			return fmt.Errorf("album %d has no cover: %w", id, errs.ErrNotFound)
		}
		oldCover = album.CoverKey

		album.CoverKey = nil
		if _, err := s.albumRepository.Update(ctx, id, album); err != nil {
			return fmt.Errorf("update cover key: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	s.coverStorage.Discard(*oldCover)
	return nil
}
