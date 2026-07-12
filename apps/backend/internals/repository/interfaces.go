package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/models"
	auth_repo "github.com/saisrikardumpeti/odoo-hackathon-2026/internals/repository/auth-repo"
)

func NewStorageRegistry(pool *pgxpool.Pool) *StorageRegistry {
	return &StorageRegistry{
		Auth: auth_repo.NewAuthRepository(pool),
	}
}

type StorageRegistry struct {
	Auth AuthStorage
}

type AuthStorage interface {
	CreateEmployeeAndUser(ctx context.Context, name, email, passwordHash string) (*models.Employee, error)
	GetEmployeeByEmail(ctx context.Context, email string) (*models.Employee, error)
	GetEmployeeByID(ctx context.Context, id string) (*models.Employee, error)
	GetUserByEmployeeID(ctx context.Context, employeeID string) (*models.User, error)
	UpdateLastLogin(ctx context.Context, userID string) error
	StoreRefreshToken(ctx context.Context, userID, refreshTokenHash string) error
	GetUserByRefreshToken(ctx context.Context, refreshTokenHash string) (*models.User, error)
	UpdatePassword(ctx context.Context, userID, passwordHash string) error
}
