package department_repo

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/models"
)

var (
	ErrDepartmentNotFound      = errors.New("department not found")
	ErrDepartmentNameExists    = errors.New("department name already exists")
	ErrParentDepartmentMissing = errors.New("parent department not found")
	ErrHeadEmployeeMissing     = errors.New("head employee not found")
)

func (r *DepartmentRepository) List(ctx context.Context) ([]models.Department, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, parent_department_id, head_employee_id, status, created_at, updated_at
		 FROM departments ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var departments []models.Department
	for rows.Next() {
		var d models.Department
		if err := rows.Scan(&d.ID, &d.Name, &d.ParentDepartmentID, &d.HeadEmployeeID,
			&d.Status, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		departments = append(departments, d)
	}
	return departments, nil
}

func (r *DepartmentRepository) GetByID(ctx context.Context, id string) (*models.Department, error) {
	var d models.Department
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, parent_department_id, head_employee_id, status, created_at, updated_at
		 FROM departments WHERE id = $1`, id,
	).Scan(&d.ID, &d.Name, &d.ParentDepartmentID, &d.HeadEmployeeID,
		&d.Status, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrDepartmentNotFound
		}
		return nil, err
	}
	return &d, nil
}

func (r *DepartmentRepository) Create(ctx context.Context, d models.Department) (*models.Department, error) {
	if d.ParentDepartmentID != nil && *d.ParentDepartmentID != "" {
		var exists bool
		err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM departments WHERE id = $1)`, *d.ParentDepartmentID).Scan(&exists)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, ErrParentDepartmentMissing
		}
	}

	if d.HeadEmployeeID != nil && *d.HeadEmployeeID != "" {
		var exists bool
		err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM employees WHERE id = $1)`, *d.HeadEmployeeID).Scan(&exists)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, ErrHeadEmployeeMissing
		}
	}

	status := "Active"
	if d.Status != "" {
		status = d.Status
	}

	var created models.Department
	err := r.pool.QueryRow(ctx,
		`INSERT INTO departments (name, parent_department_id, head_employee_id, status)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, name, parent_department_id, head_employee_id, status, created_at, updated_at`,
		d.Name, d.ParentDepartmentID, d.HeadEmployeeID, status,
	).Scan(&created.ID, &created.Name, &created.ParentDepartmentID, &created.HeadEmployeeID,
		&created.Status, &created.CreatedAt, &created.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrDepartmentNameExists
		}
		return nil, err
	}
	return &created, nil
}

func (r *DepartmentRepository) Update(ctx context.Context, d models.Department) (*models.Department, error) {
	if d.ParentDepartmentID != nil && *d.ParentDepartmentID != "" {
		var exists bool
		err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM departments WHERE id = $1)`, *d.ParentDepartmentID).Scan(&exists)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, ErrParentDepartmentMissing
		}
	}

	if d.HeadEmployeeID != nil && *d.HeadEmployeeID != "" {
		var exists bool
		err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM employees WHERE id = $1)`, *d.HeadEmployeeID).Scan(&exists)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, ErrHeadEmployeeMissing
		}
	}

	var updated models.Department
	err := r.pool.QueryRow(ctx,
		`UPDATE departments SET name = $1, parent_department_id = $2, head_employee_id = $3
		 WHERE id = $4
		 RETURNING id, name, parent_department_id, head_employee_id, status, created_at, updated_at`,
		d.Name, d.ParentDepartmentID, d.HeadEmployeeID, d.ID,
	).Scan(&updated.ID, &updated.Name, &updated.ParentDepartmentID, &updated.HeadEmployeeID,
		&updated.Status, &updated.CreatedAt, &updated.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrDepartmentNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrDepartmentNameExists
		}
		return nil, err
	}
	return &updated, nil
}

func (r *DepartmentRepository) Deactivate(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE departments SET status = 'Inactive' WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrDepartmentNotFound
	}
	return nil
}

func (r *DepartmentRepository) GetActiveEmployeeCount(ctx context.Context, departmentID string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM employees WHERE department_id = $1 AND status = 'Active'`,
		departmentID,
	).Scan(&count)
	return count, err
}
