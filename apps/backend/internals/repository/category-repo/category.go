package category_repo

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/models"
)

var (
	ErrCategoryNotFound = errors.New("category not found")
	ErrCategoryNameExists = errors.New("category name already exists")
)

func (r *CategoryRepository) List(ctx context.Context) ([]models.AssetCategory, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, custom_fields, created_at, updated_at
		 FROM asset_categories ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.AssetCategory
	for rows.Next() {
		var c models.AssetCategory
		var rawFields []byte
		if err := rows.Scan(&c.ID, &c.Name, &rawFields, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(rawFields, &c.CustomFields); err != nil {
			c.CustomFields = make(map[string]interface{})
		}
		categories = append(categories, c)
	}
	return categories, nil
}

func (r *CategoryRepository) GetByID(ctx context.Context, id string) (*models.AssetCategory, error) {
	var c models.AssetCategory
	var rawFields []byte
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, custom_fields, created_at, updated_at
		 FROM asset_categories WHERE id = $1`, id,
	).Scan(&c.ID, &c.Name, &rawFields, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCategoryNotFound
		}
		return nil, err
	}
	if err := json.Unmarshal(rawFields, &c.CustomFields); err != nil {
		c.CustomFields = make(map[string]interface{})
	}
	return &c, nil
}

func (r *CategoryRepository) Create(ctx context.Context, c models.AssetCategory) (*models.AssetCategory, error) {
	rawFields, err := json.Marshal(c.CustomFields)
	if err != nil {
		return nil, err
	}
	if rawFields == nil {
		rawFields = []byte("{}")
	}

	var created models.AssetCategory
	var resultFields []byte
	err = r.pool.QueryRow(ctx,
		`INSERT INTO asset_categories (name, custom_fields)
		 VALUES ($1, $2)
		 RETURNING id, name, custom_fields, created_at, updated_at`,
		c.Name, rawFields,
	).Scan(&created.ID, &created.Name, &resultFields, &created.CreatedAt, &created.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrCategoryNameExists
		}
		return nil, err
	}
	if err := json.Unmarshal(resultFields, &created.CustomFields); err != nil {
		created.CustomFields = make(map[string]interface{})
	}
	return &created, nil
}

func (r *CategoryRepository) Update(ctx context.Context, c models.AssetCategory) (*models.AssetCategory, error) {
	rawFields, err := json.Marshal(c.CustomFields)
	if err != nil {
		return nil, err
	}

	var updated models.AssetCategory
	var resultFields []byte
	err = r.pool.QueryRow(ctx,
		`UPDATE asset_categories SET name = $1, custom_fields = $2
		 WHERE id = $3
		 RETURNING id, name, custom_fields, created_at, updated_at`,
		c.Name, rawFields, c.ID,
	).Scan(&updated.ID, &updated.Name, &resultFields, &updated.CreatedAt, &updated.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCategoryNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrCategoryNameExists
		}
		return nil, err
	}
	if err := json.Unmarshal(resultFields, &updated.CustomFields); err != nil {
		updated.CustomFields = make(map[string]interface{})
	}
	return &updated, nil
}