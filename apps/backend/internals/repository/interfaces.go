package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/models"
	activity_log_repo "github.com/saisrikardumpeti/odoo-hackathon-2026/internals/repository/activity-log-repo"
	allocation_repo "github.com/saisrikardumpeti/odoo-hackathon-2026/internals/repository/allocation-repo"
	asset_repo "github.com/saisrikardumpeti/odoo-hackathon-2026/internals/repository/asset-repo"
	auth_repo "github.com/saisrikardumpeti/odoo-hackathon-2026/internals/repository/auth-repo"
	category_repo "github.com/saisrikardumpeti/odoo-hackathon-2026/internals/repository/category-repo"
	department_repo "github.com/saisrikardumpeti/odoo-hackathon-2026/internals/repository/department-repo"
	employee_repo "github.com/saisrikardumpeti/odoo-hackathon-2026/internals/repository/employee-repo"
	notification_repo "github.com/saisrikardumpeti/odoo-hackathon-2026/internals/repository/notification-repo"
	transfer_repo "github.com/saisrikardumpeti/odoo-hackathon-2026/internals/repository/transfer-repo"
)

func NewStorageRegistry(pool *pgxpool.Pool) *StorageRegistry {
	return &StorageRegistry{
		Auth:         auth_repo.NewAuthRepository(pool),
		Department:   department_repo.NewDepartmentRepository(pool),
		Category:     category_repo.NewCategoryRepository(pool),
		Employee:     employee_repo.NewEmployeeRepository(pool),
		ActivityLog:  activity_log_repo.NewActivityLogRepository(pool),
		Asset:        asset_repo.NewAssetRepository(pool),
		Allocation:   allocation_repo.NewAllocationRepository(pool),
		Transfer:     transfer_repo.NewTransferRepository(pool),
		Notification: notification_repo.NewNotificationRepository(pool),
	}
}

type StorageRegistry struct {
	Auth         AuthStorage
	Department   DepartmentStorage
	Category     AssetCategoryStorage
	Employee     EmployeeStorage
	ActivityLog  ActivityLogStorage
	Asset        AssetStorage
	Allocation   AllocationStorage
	Transfer     TransferStorage
	Notification NotificationStorage
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

type AssetStorage interface {
	Create(ctx context.Context, asset models.Asset) (*models.Asset, error)
	CreateStatusHistory(ctx context.Context, h models.AssetStatusHistory) error
	List(ctx context.Context, filters asset_repo.AssetListFilters) (*asset_repo.AssetListResult, error)
	GetByID(ctx context.Context, id string) (*models.AssetDetail, error)
	GetHistory(ctx context.Context, assetID string) ([]models.HistoryEvent, error)
	CreateDocument(ctx context.Context, doc models.AssetDocument) (*models.AssetDocument, error)
}

type ActivityLogStorage interface {
	Create(ctx context.Context, entry models.ActivityLog) error
}

type AllocationStorage interface {
	Create(ctx context.Context, a models.Allocation) (*models.Allocation, error)
	GetByID(ctx context.Context, id string) (*models.AllocationDetail, error)
	GetActiveByAssetID(ctx context.Context, assetID string) (*models.AllocationDetail, error)
	UpdateStatus(ctx context.Context, id, status string, returnedAt *time.Time, returnConditionNotes *string) error
	ListOverdue(ctx context.Context) ([]models.AllocationDetail, error)
	ListByEmployee(ctx context.Context, employeeID string) ([]models.AllocationDetail, error)
	UpdateAssetStatusTx(ctx context.Context, allocationID, assetID, status string, changedByID *string, reason string) error
	UpdateAssetHolder(ctx context.Context, assetID string, employeeID *string, departmentID *string) error
}

type TransferStorage interface {
	Create(ctx context.Context, t models.TransferRequest) (*models.TransferRequest, error)
	GetByID(ctx context.Context, id string) (*models.TransferRequestDetail, error)
	ListPending(ctx context.Context) ([]models.TransferRequestDetail, error)
	UpdateStatus(ctx context.Context, id, status string, approvedBy *string, approvedAt *time.Time, rejectedReason *string) error
}

type NotificationStorage interface {
	Create(ctx context.Context, n models.Notification) error
	ListByEmployee(ctx context.Context, employeeID string, unreadOnly bool) ([]models.Notification, error)
	MarkRead(ctx context.Context, id string) error
}