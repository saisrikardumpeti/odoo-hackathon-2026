package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/models"
	activity_log_repo "github.com/saisrikardumpeti/odoo-hackathon-2026/internals/repository/activity-log-repo"
	auth_repo "github.com/saisrikardumpeti/odoo-hackathon-2026/internals/repository/auth-repo"
	category_repo "github.com/saisrikardumpeti/odoo-hackathon-2026/internals/repository/category-repo"
	department_repo "github.com/saisrikardumpeti/odoo-hackathon-2026/internals/repository/department-repo"
	employee_repo "github.com/saisrikardumpeti/odoo-hackathon-2026/internals/repository/employee-repo"
)

func NewStorageRegistry(pool *pgxpool.Pool) *StorageRegistry {
	return &StorageRegistry{
		Auth:        auth_repo.NewAuthRepository(pool),
		Department:  department_repo.NewDepartmentRepository(pool),
		Category:    category_repo.NewCategoryRepository(pool),
		Employee:    employee_repo.NewEmployeeRepository(pool),
		ActivityLog: activity_log_repo.NewActivityLogRepository(pool),
	}
}

type StorageRegistry struct {
	Auth        AuthStorage
	Department  DepartmentStorage
	Category    AssetCategoryStorage
	Employee    EmployeeStorage
	ActivityLog ActivityLogStorage
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

type DepartmentStorage interface {
	List(ctx context.Context) ([]models.Department, error)
	GetByID(ctx context.Context, id string) (*models.Department, error)
	Create(ctx context.Context, d models.Department) (*models.Department, error)
	Update(ctx context.Context, d models.Department) (*models.Department, error)
	Deactivate(ctx context.Context, id string) error
	GetActiveEmployeeCount(ctx context.Context, departmentID string) (int, error)
}

type AssetCategoryStorage interface {
	List(ctx context.Context) ([]models.AssetCategory, error)
	GetByID(ctx context.Context, id string) (*models.AssetCategory, error)
	Create(ctx context.Context, c models.AssetCategory) (*models.AssetCategory, error)
	Update(ctx context.Context, c models.AssetCategory) (*models.AssetCategory, error)
}

type EmployeeStorage interface {
	List(ctx context.Context, departmentID, role, status string) ([]models.Employee, error)
	GetByID(ctx context.Context, id string) (*models.Employee, error)
	Update(ctx context.Context, e models.Employee) (*models.Employee, error)
	UpdateRole(ctx context.Context, id, role string) error
}

type ActivityLogStorage interface {
	Create(ctx context.Context, entry models.ActivityLog) error
}
