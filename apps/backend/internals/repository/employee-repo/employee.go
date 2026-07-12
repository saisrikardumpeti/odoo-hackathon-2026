package employee_repo

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/models"
)

var (
	ErrEmployeeNotFound = errors.New("employee not found")
)

func (r *EmployeeRepository) List(ctx context.Context, departmentID, role, status string) ([]models.Employee, error) {
	query := `SELECT id, name, email, department_id, role, status, created_at, updated_at FROM employees WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if departmentID != "" {
		query += fmt.Sprintf(" AND department_id = $%d", argIdx)
		args = append(args, departmentID)
		argIdx++
	}
	if role != "" {
		query += fmt.Sprintf(" AND role = $%d", argIdx)
		args = append(args, role)
		argIdx++
	}
	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}

	query += " ORDER BY name"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var employees []models.Employee
	for rows.Next() {
		var e models.Employee
		if err := rows.Scan(&e.ID, &e.Name, &e.Email, &e.DepartmentID,
			&e.Role, &e.Status, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		employees = append(employees, e)
	}
	return employees, nil
}

func (r *EmployeeRepository) GetByID(ctx context.Context, id string) (*models.Employee, error) {
	var e models.Employee
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, email, department_id, role, status, created_at, updated_at
		 FROM employees WHERE id = $1`, id,
	).Scan(&e.ID, &e.Name, &e.Email, &e.DepartmentID,
		&e.Role, &e.Status, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrEmployeeNotFound
		}
		return nil, err
	}
	return &e, nil
}

func (r *EmployeeRepository) Update(ctx context.Context, e models.Employee) (*models.Employee, error) {
	setClauses := []string{}
	args := []interface{}{}
	argIdx := 1

	if e.Name != "" {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, e.Name)
		argIdx++
	}
	if e.DepartmentID != nil {
		setClauses = append(setClauses, fmt.Sprintf("department_id = $%d", argIdx))
		args = append(args, *e.DepartmentID)
		argIdx++
	}
	if e.Status != "" {
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, e.Status)
		argIdx++
	}

	if len(setClauses) == 0 {
		return r.GetByID(ctx, e.ID)
	}

	query := fmt.Sprintf(
		`UPDATE employees SET %s WHERE id = $%d
		 RETURNING id, name, email, department_id, role, status, created_at, updated_at`,
		strings.Join(setClauses, ", "), argIdx,
	)
	args = append(args, e.ID)

	var updated models.Employee
	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&updated.ID, &updated.Name, &updated.Email, &updated.DepartmentID,
		&updated.Role, &updated.Status, &updated.CreatedAt, &updated.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrEmployeeNotFound
		}
		return nil, err
	}
	return &updated, nil
}

func (r *EmployeeRepository) UpdateRole(ctx context.Context, id, role string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE employees SET role = $1 WHERE id = $2`, role, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrEmployeeNotFound
	}
	return nil
}
