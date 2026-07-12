package auth_repo

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/models"
)

var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrEmployeeNotFound   = errors.New("employee not found")
	ErrUserNotFound       = errors.New("user not found")
)

func (r *AuthRepository) CreateEmployeeAndUser(ctx context.Context, name, email, passwordHash string) (*models.Employee, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var employee models.Employee
	err = tx.QueryRow(ctx,
		`INSERT INTO employees (name, email, role, status)
		 VALUES ($1, $2, 'Employee', 'Active')
		 RETURNING id, name, email, department_id, role, status, created_at, updated_at`,
		name, email,
	).Scan(&employee.ID, &employee.Name, &employee.Email, &employee.DepartmentID,
		&employee.Role, &employee.Status, &employee.CreatedAt, &employee.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrEmailAlreadyExists
		}
		return nil, err
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO users (employee_id, password_hash) VALUES ($1, $2)`,
		employee.ID, passwordHash,
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &employee, nil
}

func (r *AuthRepository) GetEmployeeByEmail(ctx context.Context, email string) (*models.Employee, error) {
	var e models.Employee
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, email, department_id, role, status, created_at, updated_at
		 FROM employees WHERE email = $1`,
		email,
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

func (r *AuthRepository) GetEmployeeByID(ctx context.Context, id string) (*models.Employee, error) {
	var e models.Employee
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, email, department_id, role, status, created_at, updated_at
		 FROM employees WHERE id = $1`,
		id,
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

func (r *AuthRepository) GetUserByEmployeeID(ctx context.Context, employeeID string) (*models.User, error) {
	var u models.User
	var refreshTokenHash *string
	var lastLoginAt *time.Time

	err := r.pool.QueryRow(ctx,
		`SELECT id, employee_id, password_hash, refresh_token_hash, last_login_at, created_at, updated_at
		 FROM users WHERE employee_id = $1`,
		employeeID,
	).Scan(&u.ID, &u.EmployeeID, &u.PasswordHash, &refreshTokenHash,
		&lastLoginAt, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	u.RefreshTokenHash = refreshTokenHash
	u.LastLoginAt = lastLoginAt
	return &u, nil
}

func (r *AuthRepository) UpdateLastLogin(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET last_login_at = now() WHERE id = $1`,
		userID,
	)
	return err
}

func (r *AuthRepository) StoreRefreshToken(ctx context.Context, userID, refreshTokenHash string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET refresh_token_hash = $1 WHERE id = $2`,
		refreshTokenHash, userID,
	)
	return err
}

func (r *AuthRepository) GetUserByRefreshToken(ctx context.Context, refreshTokenHash string) (*models.User, error) {
	var u models.User
	var lastLoginAt *time.Time

	err := r.pool.QueryRow(ctx,
		`SELECT id, employee_id, password_hash, refresh_token_hash, last_login_at, created_at, updated_at
		 FROM users WHERE refresh_token_hash = $1`,
		refreshTokenHash,
	).Scan(&u.ID, &u.EmployeeID, &u.PasswordHash, &u.RefreshTokenHash,
		&lastLoginAt, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	u.LastLoginAt = lastLoginAt
	return &u, nil
}

func (r *AuthRepository) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET password_hash = $1 WHERE id = $2`,
		passwordHash, userID,
	)
	return err
}