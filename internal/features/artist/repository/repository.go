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

func (r *Repository) List(ctx context.Context, limit *int, offset *int) ([]domain.Artist, error) {
	queryCtx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	sqlQuery := `
		SELECT id, name, country
		FROM mc.artist
		ORDER BY id ASC
		LIMIT $1
		OFFSET $2;
	`

	rows, err := r.pool.Querier(ctx).Query(queryCtx, sqlQuery, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("pool query: %w", err)
	}
	defer rows.Close()

	var artistModels []ArtistModel
	for rows.Next() {
		var artistModel ArtistModel
		if err := rows.Scan(&artistModel.ID, &artistModel.Name, &artistModel.Country); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		artistModels = append(artistModels, artistModel)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}

	artistDomains := make([]domain.Artist, 0, len(artistModels))
	for _, m := range artistModels {
		artistDomains = append(artistDomains, m.ToDomain())
	}

	return artistDomains, nil
}

func (r *Repository) GetByID(ctx context.Context, id int) (domain.Artist, error) {
	return r.getByID(ctx, id, false)
}

func (r *Repository) GetByIDForUpdate(ctx context.Context, id int) (domain.Artist, error) {
	return r.getByID(ctx, id, true)
}

func (r *Repository) getByID(ctx context.Context, id int, forUpdate bool) (domain.Artist, error) {
	queryCtx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	sqlQuery := `
		SELECT id, name, country
		FROM mc.artist
		WHERE id = $1
		%s;
	`

	if forUpdate {
		sqlQuery = fmt.Sprintf(sqlQuery, "FOR UPDATE")
	} else {
		sqlQuery = fmt.Sprintf(sqlQuery, "")
	}

	row := r.pool.Querier(ctx).QueryRow(queryCtx, sqlQuery, id)

	var artistModel ArtistModel
	if err := row.Scan(&artistModel.ID, &artistModel.Name, &artistModel.Country); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Artist{}, fmt.Errorf(
				"artist with id: %d: %w",
				id,
				errs.ErrNotFound,
			)
		}
		return domain.Artist{}, fmt.Errorf("scan row: %w", err)
	}

	return artistModel.ToDomain(), nil
}

func (r *Repository) Create(ctx context.Context, artist domain.Artist) (domain.Artist, error) {
	queryCtx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	sqlQuery := `
		INSERT INTO mc.artist(name, country)
		VALUES ($1, $2)
		RETURNING id, name, country;
	`

	row := r.pool.Querier(ctx).QueryRow(queryCtx, sqlQuery, artist.Name, artist.Country)

	var artistModel ArtistModel
	if err := row.Scan(&artistModel.ID, &artistModel.Name, &artistModel.Country); err != nil {
		return domain.Artist{}, fmt.Errorf("scan row: %w", err)
	}

	return artistModel.ToDomain(), nil
}

func (r *Repository) Update(ctx context.Context, id int, artist domain.Artist) (domain.Artist, error) {
	queryCtx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	sqlQuery := `
		UPDATE mc.artist
		SET
			name = $1,
			country = $2
		WHERE id = $3
		RETURNING id, name, country;
	`

	row := r.pool.Querier(ctx).QueryRow(queryCtx, sqlQuery, artist.Name, artist.Country, id)

	var artistModel ArtistModel
	if err := row.Scan(&artistModel.ID, &artistModel.Name, &artistModel.Country); err != nil {
		return domain.Artist{}, fmt.Errorf("scan row: %w", err)
	}

	return artistModel.ToDomain(), nil
}

func (r *Repository) Delete(ctx context.Context, id int) error {
	queryCtx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	sqlQuery := `
		DELETE FROM mc.artist
		WHERE id = $1;
	`

	cmdTag, err := r.pool.Querier(ctx).Exec(queryCtx, sqlQuery, id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == db.CodeForeignKeyViolation {
			return fmt.Errorf(
				"artist with id: %d has albums and cannot be deleted: %w",
				id,
				errs.ErrConflict,
			)
		}

		return fmt.Errorf("pool exec: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf(
			"artist with id: %d: %w",
			id,
			errs.ErrNotFound,
		)
	}

	return nil
}
