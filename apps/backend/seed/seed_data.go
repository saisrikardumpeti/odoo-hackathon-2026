package seed

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

var testPasswordHash string

func init() {
	h, err := bcrypt.GenerateFromPassword([]byte("test123"), bcrypt.DefaultCost)
	if err != nil {
		panic(fmt.Sprintf("failed to hash test password: %v", err))
	}
	testPasswordHash = string(h)
}

func SeedTestData(ctx context.Context, pool *pgxpool.Pool) error {
	var count int
	err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM departments`).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check existing data: %w", err)
	}
	if count > 0 {
		log.Println("Test data already exists, skipping seed")
		return nil
	}

	log.Println("Seeding test data...")

	// Departments
	engID := execScalar(ctx, pool, `INSERT INTO departments (name) VALUES ($1) RETURNING id`, "Engineering")
	mktID := execScalar(ctx, pool, `INSERT INTO departments (name) VALUES ($1) RETURNING id`, "Marketing")
	execScalar(ctx, pool, `INSERT INTO departments (name) VALUES ($1) RETURNING id`, "Finance")
	opsID := execScalar(ctx, pool, `INSERT INTO departments (name) VALUES ($1) RETURNING id`, "Operations")

	// Asset categories
	roomCatID := execScalar(ctx, pool,
		`INSERT INTO asset_categories (name, custom_fields) VALUES ($1, $2::jsonb) RETURNING id`,
		"Meeting Rooms", `{"capacity": "number", "has_projector": "boolean"}`)
	vehCatID := execScalar(ctx, pool,
		`INSERT INTO asset_categories (name, custom_fields) VALUES ($1, $2::jsonb) RETURNING id`,
		"Vehicles", `{"license_plate": "text", "fuel_type": "text"}`)
	equipCatID := execScalar(ctx, pool,
		`INSERT INTO asset_categories (name, custom_fields) VALUES ($1, $2::jsonb) RETURNING id`,
		"Equipment", `{"warranty_years": "number"}`)

	// Employees (with user accounts)
	aliceID := insEmployee(ctx, pool, "Alice Johnson", "alice@assetflow.local", engID, "DepartmentHead")
	bobID := insEmployee(ctx, pool, "Bob Smith", "bob@assetflow.local", engID, "Employee")
	carolID := insEmployee(ctx, pool, "Carol Davis", "carol@assetflow.local", mktID, "DepartmentHead")
	insEmployee(ctx, pool, "David Lee", "david@assetflow.local", mktID, "Employee")
	insEmployee(ctx, pool, "Eve Chen", "eve@assetflow.local", opsID, "AssetManager")
	insEmployee(ctx, pool, "Frank Brown", "frank@assetflow.local", opsID, "Employee")

	// Set department heads
	execVoid(ctx, pool, `UPDATE departments SET head_employee_id = $1 WHERE id = $2`, aliceID, engID)
	execVoid(ctx, pool, `UPDATE departments SET head_employee_id = $1 WHERE id = $2`, carolID, mktID)

	// Bookable assets (is_bookable = true)
	rB1 := insBookable(ctx, pool, "Conference Room A", roomCatID, `{"location":"10th floor","capacity":12}`)
	rB2 := insBookable(ctx, pool, "Conference Room B", roomCatID, `{"location":"10th floor","capacity":8}`)
	insBookable(ctx, pool, "Board Room", roomCatID, `{"location":"5th floor","capacity":20,"has_projector":true}`)
	v1 := insBookable(ctx, pool, "Toyota Camry", vehCatID, `{"location":"Parking Lot A","license_plate":"ABC-1234","fuel_type":"Hybrid"}`)
	insBookable(ctx, pool, "Ford Transit", vehCatID, `{"location":"Parking Lot B","license_plate":"XYZ-5678","fuel_type":"Diesel"}`)
	insBookable(ctx, pool, "Projector Kit", equipCatID, `{"location":"AV Closet Floor 10","warranty_years":3}`)
	insBookable(ctx, pool, "Sound System", equipCatID, `{"location":"AV Closet Floor 5","warranty_years":5}`)

	// Non-bookable assets (is_bookable = false)
	insNonBookable(ctx, pool, "Dell Latitude 5540", equipCatID, "SN-LAP-001", aliceID)
	insNonBookable(ctx, pool, "MacBook Pro 16", equipCatID, "SN-LAP-002", bobID)

	// Sample bookings
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	tomorrow := today.AddDate(0, 0, 1)

	insBooking(ctx, pool, rB2, aliceID, today.Add(9*time.Hour), today.Add(10*time.Hour), "Weekly sync")
	insBooking(ctx, pool, rB1, bobID, today.Add(13*time.Hour), today.Add(15*time.Hour), "Sprint planning")
	insBooking(ctx, pool, v1, carolID, tomorrow.Add(10*time.Hour), tomorrow.Add(12*time.Hour), "Client visit")

	log.Println("Test data seeded successfully")
	log.Println("Sample logins (password: test123):")
	log.Println("  alice@assetflow.local — Engineering DepartmentHead")
	log.Println("  bob@assetflow.local — Engineering Employee")
	log.Println("  carol@assetflow.local — Marketing DepartmentHead")
	log.Println("  eve@assetflow.local — Operations AssetManager")
	log.Println("  frank@assetflow.local — Operations Employee")
	log.Println("  admin@assetflow.local / admin123 — System Admin")
	return nil
}

func execScalar(ctx context.Context, pool *pgxpool.Pool, query string, args ...interface{}) string {
	var id string
	if err := pool.QueryRow(ctx, query, args...).Scan(&id); err != nil {
		panic(fmt.Sprintf("seed query failed: %v\nSQL: %s", err, query))
	}
	return id
}

func execVoid(ctx context.Context, pool *pgxpool.Pool, query string, args ...interface{}) {
	if _, err := pool.Exec(ctx, query, args...); err != nil {
		panic(fmt.Sprintf("seed exec failed: %v\nSQL: %s", err, query))
	}
}

func insEmployee(ctx context.Context, pool *pgxpool.Pool, name, email, deptID, role string) string {
	eid := execScalar(ctx, pool,
		`INSERT INTO employees (name, email, department_id, role, status)
		 VALUES ($1, $2, $3, $4, 'Active') RETURNING id`,
		name, email, deptID, role)
	execVoid(ctx, pool,
		`INSERT INTO users (employee_id, password_hash) VALUES ($1, $2)`, eid, testPasswordHash)
	return eid
}

func insBookable(ctx context.Context, pool *pgxpool.Pool, name, categoryID, customFields string) string {
	return execScalar(ctx, pool,
		`INSERT INTO assets (name, category_id, is_bookable, status, custom_fields)
		 VALUES ($1, $2, true, 'Available', $3::jsonb) RETURNING id`,
		name, categoryID, customFields)
}

func insNonBookable(ctx context.Context, pool *pgxpool.Pool, name, categoryID, serial, employeeID string) {
	execScalar(ctx, pool,
		`INSERT INTO assets (name, category_id, is_bookable, status, serial_number, current_holder_employee_id)
		 VALUES ($1, $2, false, 'Allocated', $3, $4) RETURNING id`,
		name, categoryID, serial, employeeID)
}

func insBooking(ctx context.Context, pool *pgxpool.Pool, resourceID, employeeID string, start, end time.Time, purpose string) {
	execScalar(ctx, pool,
		`INSERT INTO bookings (resource_asset_id, booked_by_employee_id, start_time, end_time, purpose, status)
		 VALUES ($1, $2, $3, $4, $5, 'Upcoming') RETURNING id`,
		resourceID, employeeID, start, end, purpose)
}


