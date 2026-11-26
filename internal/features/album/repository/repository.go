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

func (r *Repository) List(ctx context.Context, artistID *int, limit *int, offset *int) ([]domain.Album, error) {
	queryCtx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	sqlQuery := `
		SELECT id, artist_id, title, type, release_date, cover_key
		FROM mc.album
		%s
		ORDER BY id ASC
		LIMIT $1
		OFFSET $2;
	`

	args := []any{limit, offset}

	if artistID != nil {
		sqlQuery = fmt.Sprintf(sqlQuery, "WHERE artist_id = $3")
		args = append(args, artistID)
	} else {
		sqlQuery = fmt.Sprintf(sqlQuery, "")
	}

	rows, err := r.pool.Querier(ctx).Query(queryCtx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("pool query: %w", err)
	}
	defer rows.Close()

	var albumModels []AlbumModel
	for rows.Next() {
		var albumModel AlbumModel
		if err := rows.Scan(
			&albumModel.ID,
			&albumModel.ArtistID,
			&albumModel.Title,
			&albumModel.Type,
			&albumModel.ReleaseDate,
			&albumModel.CoverKey,
		); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		albumModels = append(albumModels, albumModel)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
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
		SELECT id, artist_id, title, type, release_date, cover_key
		FROM mc.album
		WHERE id = $1
		%s;
	`

	if forUpdate {
		sqlQuery = fmt.Sprintf(sqlQuery, "FOR UPDATE")
	} else {
		sqlQuery = fmt.Sprintf(sqlQuery, "")
	}

	row := r.pool.Querier(ctx).QueryRow(queryCtx, sqlQuery, id)

	var albumModel AlbumModel
	if err := row.Scan(
		&albumModel.ID,
		&albumModel.ArtistID,
		&albumModel.Title,
		&albumModel.Type,
		&albumModel.ReleaseDate,
		&albumModel.CoverKey,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Album{}, fmt.Errorf(
				"album with id: %d: %w",
				id,
				errs.ErrNotFound,
			)
		}
		return domain.Album{}, fmt.Errorf("scan row: %w", err)
	}

	return albumModel.ToDomain(), nil
}

func (r *Repository) Create(ctx context.Context, album domain.Album) (domain.Album, error) {
	queryCtx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	sqlQuery := `
		INSERT INTO mc.album(artist_id, title, type, release_date, cover_key)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, artist_id, title, type, release_date, cover_key;
	`

	row := r.pool.Querier(ctx).QueryRow(
		queryCtx,
		sqlQuery,
		album.ArtistID,
		album.Title,
		album.Type,
		album.ReleaseDate,
		album.CoverKey,
	)

	var albumModel AlbumModel
	if err := row.Scan(
		&albumModel.ID,
		&albumModel.ArtistID,
		&albumModel.Title,
		&albumModel.Type,
		&albumModel.ReleaseDate,
		&albumModel.CoverKey,
	); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == db.CodeForeignKeyViolation {
			return domain.Album{}, fmt.Errorf(
				"artist with id: %d: does not exist: %w",
				album.ArtistID,
				errs.ErrUnprocessableEntity,
			)
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
			release_date = $4,
			cover_key = $5
		WHERE id = $6
		RETURNING id, artist_id, title, type, release_date, cover_key;
	`

	row := r.pool.Querier(ctx).QueryRow(
		queryCtx,
		sqlQuery,
		album.ArtistID,
		album.Title,
		album.Type,
		album.ReleaseDate,
		album.CoverKey,
		id,
	)

	var albumModel AlbumModel
	if err := row.Scan(
		&albumModel.ID,
		&albumModel.ArtistID,
		&albumModel.Title,
		&albumModel.Type,
		&albumModel.ReleaseDate,
		&albumModel.CoverKey,
	); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == db.CodeForeignKeyViolation {
			return domain.Album{}, fmt.Errorf(
				"artist with id: %d: does not exist: %w",
				album.ArtistID,
				errs.ErrUnprocessableEntity,
			)
		}
		return domain.Album{}, fmt.Errorf("scan row: %w", err)
	}

	return albumModel.ToDomain(), nil
}

func (r *Repository) Delete(ctx context.Context, id int) (*string, error) {
	queryCtx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	sqlQuery := `
		DELETE FROM mc.album
		WHERE id = $1
		RETURNING cover_key;
	`

	row := r.pool.Querier(ctx).QueryRow(queryCtx, sqlQuery, id)

	var coverKey *string
	if err := row.Scan(&coverKey); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf(
				"album with id: %d: %w",
				id,
				errs.ErrNotFound,
			)
		}
		return nil, fmt.Errorf("scan row: %w", err)
	}

	return coverKey, nil
}
