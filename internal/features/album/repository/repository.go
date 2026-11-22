package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nseve/music-catalog/internal/core/db"
	"github.com/nseve/music-catalog/internal/core/domain"
	errs "github.com/nseve/music-catalog/internal/core/errors"
)

type Repository struct {
	pool *db.Pool
}

func New(pool *db.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}

func (r *Repository) GetAll(ctx context.Context) ([]domain.Album, error) {
	queryCtx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	sqlQuery := `
		SELECT id, artist_id, title, type, release_date, cover_path
		FROM mc.album
		ORDER BY id ASC;
	`

	rows, err := r.pool.Querier(ctx).Query(queryCtx, sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("pool query: %w", err)
	}

	var albumModels []AlbumModel
	for rows.Next() {
		var albumModel AlbumModel
		if err := rows.Scan(
			&albumModel.ID,
			&albumModel.ArtistID,
			&albumModel.Title,
			&albumModel.Type,
			&albumModel.ReleaseDate,
			&albumModel.CoverPath,
		); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		albumModels = append(albumModels, albumModel)
	}

	albumDomains := make([]domain.Album, 0, len(albumModels))
	for _, m := range albumModels {
		albumDomains = append(albumDomains, m.ToDomain())
	}

	return albumDomains, nil
}

func (r *Repository) GetByID(ctx context.Context, id int) (domain.Album, error) {
	return r.getByID(ctx, id, false)
}

func (r *Repository) GetByIDForUpdate(ctx context.Context, id int) (domain.Album, error) {
	return r.getByID(ctx, id, true)
}

func (r *Repository) getByID(ctx context.Context, id int, forUpdate bool) (domain.Album, error) {
	queryCtx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	sqlQuery := `
		SELECT id, artist_id, title, type, release_date, cover_path
		FROM mc.album
		WHERE id = $1
	`
	if forUpdate {
		sqlQuery += ` FOR UPDATE`
	}

	row := r.pool.Querier(ctx).QueryRow(queryCtx, sqlQuery, id)

	var albumModel AlbumModel
	if err := row.Scan(
		&albumModel.ID,
		&albumModel.ArtistID,
		&albumModel.Title,
		&albumModel.Type,
		&albumModel.ReleaseDate,
		&albumModel.CoverPath,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Album{}, fmt.Errorf("album with id: %d: %w", id, errs.ErrNotFound)
		}
		return domain.Album{}, fmt.Errorf("scan row: %w", err)
	}

	return albumModel.ToDomain(), nil
}

func (r *Repository) Create(ctx context.Context, album domain.Album) (domain.Album, error) {
	queryCtx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	sqlQuery := `
		INSERT INTO mc.album(artist_id, title, type, release_date, cover_path)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, artist_id, title, type, release_date, cover_path;
	`

	row := r.pool.Querier(ctx).QueryRow(
		queryCtx,
		sqlQuery,
		album.ArtistID,
		album.Title,
		album.Type,
		album.ReleaseDate,
		album.CoverPath,
	)

	var albumModel AlbumModel
	if err := row.Scan(
		&albumModel.ID,
		&albumModel.ArtistID,
		&albumModel.Title,
		&albumModel.Type,
		&albumModel.ReleaseDate,
		&albumModel.CoverPath,
	); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == db.CodeForeignKeyViolation {
			return domain.Album{}, fmt.Errorf("artist with id: %d: %w", album.ArtistID, errs.ErrNotFound)
		}
		return domain.Album{}, fmt.Errorf("scan row: %w", err)
	}

	return albumModel.ToDomain(), nil
}

func (r *Repository) Update(ctx context.Context, id int, album domain.Album) (domain.Album, error) {
	queryCtx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	sqlQuery := `
		UPDATE mc.album
		SET
			artist_id = $1,
			title = $2,
			type = $3,
			release_date = $4
		WHERE id = $5
		RETURNING id, artist_id, title, type, release_date, cover_path;
	`

	row := r.pool.Querier(ctx).QueryRow(
		queryCtx,
		sqlQuery,
		album.ArtistID,
		album.Title,
		album.Type,
		album.ReleaseDate,
		id,
	)

	var albumModel AlbumModel
	if err := row.Scan(
		&albumModel.ID,
		&albumModel.ArtistID,
		&albumModel.Title,
		&albumModel.Type,
		&albumModel.ReleaseDate,
		&albumModel.CoverPath,
	); err != nil {
		return domain.Album{}, fmt.Errorf("scan row: %w", err)
	}

	return albumModel.ToDomain(), nil
}

func (r *Repository) Delete(ctx context.Context, id int) error {
	queryCtx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	sqlQuery := `
		DELETE FROM mc.album
		WHERE id = $1;
	`

	cmdTag, err := r.pool.Querier(ctx).Exec(queryCtx, sqlQuery, id)
	if err != nil {
		return fmt.Errorf("pool exec: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("album with id: %d: %w", id, errs.ErrNotFound)
	}

	return nil
}
