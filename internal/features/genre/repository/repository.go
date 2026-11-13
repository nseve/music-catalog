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

func (r *Repository) GetAll(ctx context.Context) ([]domain.Genre, error) {
	queryCtx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	sqlQuery := `
		SELECT id, name
		FROM mc.genre
		ORDER BY id ASC;
	`

	rows, err := r.pool.Querier(ctx).Query(queryCtx, sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("pool query: %w", err)
	}

	var genreModels []GenreModel
	for rows.Next() {
		var genreModel GenreModel
		if err := rows.Scan(&genreModel.ID, &genreModel.Name); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		genreModels = append(genreModels, genreModel)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}

	genreDomains := make([]domain.Genre, 0, len(genreModels))
	for _, m := range genreModels {
		genreDomains = append(genreDomains, m.ToDomain())
	}

	return genreDomains, nil
}

func (r *Repository) GetByID(ctx context.Context, id int) (domain.Genre, error) {
	return r.getByID(ctx, id, false)
}

func (r *Repository) GetByIDForUpdate(ctx context.Context, id int) (domain.Genre, error) {
	return r.getByID(ctx, id, true)
}

func (r *Repository) getByID(ctx context.Context, id int, forUpdate bool) (domain.Genre, error) {
	queryCtx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	sqlQuery := `
		SELECT id, name
		FROM mc.genre
		WHERE id = $1
	`
	if forUpdate {
		sqlQuery += ` FOR UPDATE`
	}

	row := r.pool.Querier(ctx).QueryRow(queryCtx, sqlQuery, id)

	var genreModel GenreModel
	if err := row.Scan(&genreModel.ID, &genreModel.Name); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Genre{}, fmt.Errorf("genre with id: %d: %w", id, errs.ErrNotFound)
		}
		return domain.Genre{}, fmt.Errorf("scan row: %w", err)
	}

	return genreModel.ToDomain(), nil
}

func (r *Repository) Create(ctx context.Context, genre domain.Genre) (domain.Genre, error) {
	queryCtx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	sqlQuery := `
		INSERT INTO mc.genre(name)
		VALUES ($1)
		RETURNING id, name;
	`

	row := r.pool.Querier(ctx).QueryRow(queryCtx, sqlQuery, genre.Name)

	var genreModel GenreModel
	if err := row.Scan(&genreModel.ID, &genreModel.Name); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == db.CodeUniqueViolation {
			return domain.Genre{}, fmt.Errorf("genre with name: %s already exists: %w", genre.Name, errs.ErrConflict)
		}
		return domain.Genre{}, fmt.Errorf("scan row: %w", err)
	}

	return genreModel.ToDomain(), nil
}

func (r *Repository) Update(ctx context.Context, id int, genre domain.Genre) (domain.Genre, error) {
	queryCtx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	sqlQuery := `
		UPDATE mc.genre
		SET name = $1
		WHERE id = $2
		RETURNING id, name;
	`

	row := r.pool.Querier(ctx).QueryRow(queryCtx, sqlQuery, genre.Name, genre.ID)

	var genreModel GenreModel
	if err := row.Scan(&genreModel.ID, &genreModel.Name); err != nil {
		return domain.Genre{}, fmt.Errorf("scan row: %w", err)
	}

	return genreModel.ToDomain(), nil
}

func (r *Repository) Delete(ctx context.Context, id int) error {
	queryCtx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	sqlQuery := `
		DELETE FROM mc.genre
		WHERE id = $1;
	`

	cmdTag, err := r.pool.Querier(ctx).Exec(queryCtx, sqlQuery, id)
	if err != nil {
		return fmt.Errorf("pool exec: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("genre with id: %d: %w", id, errs.ErrNotFound)
	}

	return nil
}
