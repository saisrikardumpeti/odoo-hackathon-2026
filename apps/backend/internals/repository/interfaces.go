package repository

import (
	"context"
	"database/sql"

	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/models"
	example_repo "github.com/saisrikardumpeti/odoo-hackathon-2026/internals/repository/example-repo"
)

func NewStorageRegistry(db *sql.DB) *StorageRegistry {
	return &StorageRegistry{
		Example: example_repo.NewExampleRepository(db),
	}
}

type StorageRegistry struct {
	Example ExampleStorage
}

type ExampleStorage interface {
	CreateExample(ctx context.Context, example models.Example) error
}
