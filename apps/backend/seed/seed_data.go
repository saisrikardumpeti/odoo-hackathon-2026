package seed

import (
	"context"
	"fmt"
	"log"
	"math/rand"
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

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

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
	now := time.Now().UTC()

	// ===================================================================
	// DEPARTMENTS (6 rows)
	// ===================================================================
	engID := execScalar(ctx, pool, `INSERT INTO departments (name) VALUES ($1) RETURNING id`, "Engineering")
	mktID := execScalar(ctx, pool, `INSERT INTO departments (name) VALUES ($1) RETURNING id`, "Marketing")
	execScalar(ctx, pool, `INSERT INTO departments (name) VALUES ($1) RETURNING id`, "Finance")
	opsID := execScalar(ctx, pool, `INSERT INTO departments (name) VALUES ($1) RETURNING id`, "Operations")
	hrID := execScalar(ctx, pool, `INSERT INTO departments (name) VALUES ($1) RETURNING id`, "Human Resources")
	itsID := execScalar(ctx, pool, `INSERT INTO departments (name) VALUES ($1) RETURNING id`, "IT Support")

	// ===================================================================
	// ASSET CATEGORIES (5 rows)
	// ===================================================================
	roomCatID := execScalar(ctx, pool,
		`INSERT INTO asset_categories (name, custom_fields) VALUES ($1, $2::jsonb) RETURNING id`,
		"Meeting Rooms", `{"capacity": "number", "has_projector": "boolean"}`)
	vehCatID := execScalar(ctx, pool,
		`INSERT INTO asset_categories (name, custom_fields) VALUES ($1, $2::jsonb) RETURNING id`,
		"Vehicles", `{"license_plate": "text", "fuel_type": "text"}`)
	equipCatID := execScalar(ctx, pool,
		`INSERT INTO asset_categories (name, custom_fields) VALUES ($1, $2::jsonb) RETURNING id`,
		"Equipment", `{"warranty_years": "number"}`)
	itCatID := execScalar(ctx, pool,
		`INSERT INTO asset_categories (name, custom_fields) VALUES ($1, $2::jsonb) RETURNING id`,
		"IT Equipment", `{"os": "text", "ram_gb": "number"}`)
	furnCatID := execScalar(ctx, pool,
		`INSERT INTO asset_categories (name, custom_fields) VALUES ($1, $2::jsonb) RETURNING id`,
		"Furniture", `{"material": "text", "color": "text"}`)

	// ===================================================================
	// EMPLOYEES (8 rows)
	// ===================================================================
	aliceID := insEmployee(ctx, pool, "Alice Johnson", "alice@assetflow.local", engID, "DepartmentHead")
	bobID := insEmployee(ctx, pool, "Bob Smith", "bob@assetflow.local", engID, "Employee")
	carolID := insEmployee(ctx, pool, "Carol Davis", "carol@assetflow.local", mktID, "DepartmentHead")
	davidID := insEmployee(ctx, pool, "David Lee", "david@assetflow.local", mktID, "Employee")
	eveID := insEmployee(ctx, pool, "Eve Chen", "eve@assetflow.local", opsID, "AssetManager")
	frankID := insEmployee(ctx, pool, "Frank Brown", "frank@assetflow.local", opsID, "Employee")
	graceID := insEmployee(ctx, pool, "Grace Kim", "grace@assetflow.local", hrID, "DepartmentHead")
	henryID := insEmployee(ctx, pool, "Henry Wilson", "henry@assetflow.local", itsID, "Employee")

	// Set department heads
	execVoid(ctx, pool, `UPDATE departments SET head_employee_id = $1 WHERE id = $2`, aliceID, engID)
	execVoid(ctx, pool, `UPDATE departments SET head_employee_id = $1 WHERE id = $2`, carolID, mktID)
	execVoid(ctx, pool, `UPDATE departments SET head_employee_id = $1 WHERE id = $2`, graceID, hrID)

	// ===================================================================
	// ASSETS (15 rows — mix of bookable and non-bookable)
	// ===================================================================
	rB1 := insBookable(ctx, pool, "Conference Room A", roomCatID, `{"location":"10th floor","capacity":12}`)
	rB2 := insBookable(ctx, pool, "Conference Room B", roomCatID, `{"location":"10th floor","capacity":8}`)
	boardRoom := insBookable(ctx, pool, "Board Room", roomCatID, `{"location":"5th floor","capacity":20,"has_projector":true}`)
	trainingRoom := insBookable(ctx, pool, "Training Room", roomCatID, `{"location":"3rd floor","capacity":30,"has_projector":true}`)
	v1 := insBookable(ctx, pool, "Toyota Camry", vehCatID, `{"location":"Parking Lot A","license_plate":"ABC-1234","fuel_type":"Hybrid"}`)
	v2 := insBookable(ctx, pool, "Ford Transit", vehCatID, `{"location":"Parking Lot B","license_plate":"XYZ-5678","fuel_type":"Diesel"}`)
	v3 := insBookable(ctx, pool, "Honda Civic", vehCatID, `{"location":"Parking Lot A","license_plate":"DEF-9012","fuel_type":"Petrol"}`)
	projector := insBookable(ctx, pool, "Projector Kit", equipCatID, `{"location":"AV Closet Floor 10","warranty_years":3}`)
	soundSystem := insBookable(ctx, pool, "Sound System", equipCatID, `{"location":"AV Closet Floor 5","warranty_years":5}`)

	laptop1 := insNonBookableGetID(ctx, pool, "Dell Latitude 5540", itCatID, "SN-LAP-001", strPtr(aliceID), "Allocated", timePtr(now.AddDate(-6, 0, 0)))
	laptop2 := insNonBookableGetID(ctx, pool, "MacBook Pro 16", itCatID, "SN-LAP-002", strPtr(bobID), "Allocated", timePtr(now.AddDate(-4, 0, 0)))
	laptop3 := insNonBookableGetID(ctx, pool, "Lenovo ThinkPad X1", itCatID, "SN-LAP-003", strPtr(carolID), "Allocated", timePtr(now.AddDate(-2, 0, 0)))
	monitor := insNonBookableGetID(ctx, pool, "Dell UltraSharp 27", itCatID, "SN-MON-001", strPtr(davidID), "Allocated", timePtr(now.AddDate(-1, 0, 0)))
	server := insNonBookableGetID(ctx, pool, "Dell PowerEdge R740", itCatID, "SN-SRV-001", nil, "Available", nil)
	officeChair := insNonBookableGetID(ctx, pool, "Herman Miller Aeron", furnCatID, "FUR-CHR-001", strPtr(frankID), "Allocated", timePtr(now.AddDate(-3, 0, 0)))

	// ===================================================================
	// ASSET DOCUMENTS (8 rows)
	// ===================================================================
	execVoid(ctx, pool,
		`INSERT INTO asset_documents (asset_id, url, type) VALUES ($1, $2, $3)`, laptop1, "/docs/laptop1-receipt.pdf", "document")
	execVoid(ctx, pool,
		`INSERT INTO asset_documents (asset_id, url, type) VALUES ($1, $2, $3)`, laptop2, "/docs/laptop2-receipt.pdf", "document")
	execVoid(ctx, pool,
		`INSERT INTO asset_documents (asset_id, url, type) VALUES ($1, $2, $3)`, server, "/docs/server-specs.pdf", "document")
	execVoid(ctx, pool,
		`INSERT INTO asset_documents (asset_id, url, type) VALUES ($1, $2, $3)`, v1, "/photos/toyota-camry.jpg", "photo")
	execVoid(ctx, pool,
		`INSERT INTO asset_documents (asset_id, url, type) VALUES ($1, $2, $3)`, v2, "/photos/ford-transit.jpg", "photo")
	execVoid(ctx, pool,
		`INSERT INTO asset_documents (asset_id, url, type) VALUES ($1, $2, $3)`, rB1, "/photos/conf-room-a.jpg", "photo")
	execVoid(ctx, pool,
		`INSERT INTO asset_documents (asset_id, url, type) VALUES ($1, $2, $3)`, officeChair, "/docs/chair-warranty.pdf", "document")
	execVoid(ctx, pool,
		`INSERT INTO asset_documents (asset_id, url, type) VALUES ($1, $2, $3)`, monitor, "/docs/monitor-receipt.pdf", "document")

	// ===================================================================
	// ASSET STATUS HISTORY (12 rows)
	// ===================================================================
	insStatusHistory(ctx, pool, laptop1, nil, strPtr("Allocated"), aliceID, "Initial allocation")
	insStatusHistory(ctx, pool, laptop2, nil, strPtr("Allocated"), bobID, "Initial allocation")
	insStatusHistory(ctx, pool, laptop3, nil, strPtr("Allocated"), carolID, "Initial allocation")
	insStatusHistory(ctx, pool, server, nil, strPtr("Available"), eveID, "Server provisioned")
	insStatusHistory(ctx, pool, monitor, nil, strPtr("Allocated"), davidID, "Assigned to David")
	insStatusHistory(ctx, pool, officeChair, nil, strPtr("Allocated"), frankID, "Assigned to Frank")
	insStatusHistory(ctx, pool, v1, nil, strPtr("Available"), eveID, "Vehicle registered")
	insStatusHistory(ctx, pool, projector, strPtr("Available"), strPtr("UnderMaintenance"), eveID, "Bulb replacement needed")
	insStatusHistory(ctx, pool, soundSystem, strPtr("Available"), strPtr("UnderMaintenance"), eveID, "Speaker crackling")
	insStatusHistory(ctx, pool, laptop2, strPtr("Allocated"), strPtr("Available"), bobID, "Returned by Bob")
	insStatusHistory(ctx, pool, laptop2, strPtr("Available"), strPtr("Allocated"), bobID, "Re-allocated to Bob")
	insStatusHistory(ctx, pool, trainingRoom, nil, strPtr("Available"), aliceID, "Room added")

	// ===================================================================
	// ALLOCATIONS (8 rows)
	// ===================================================================
	alloc1 := insAllocation(ctx, pool, laptop1, strPtr(aliceID), strPtr(engID), aliceID, now.AddDate(0, 0, -60), timePtr(now.AddDate(0, 0, 30)), timePtr(now.AddDate(0, 0, -5)), "Active", nil)
	alloc2 := insAllocation(ctx, pool, laptop2, strPtr(bobID), strPtr(engID), aliceID, now.AddDate(0, 0, -45), nil, nil, "Active", nil)
	alloc3 := insAllocation(ctx, pool, laptop3, strPtr(carolID), strPtr(mktID), carolID, now.AddDate(0, 0, -30), timePtr(now.AddDate(0, 0, 15)), nil, "Active", nil)
	alloc4 := insAllocation(ctx, pool, monitor, strPtr(davidID), strPtr(mktID), carolID, now.AddDate(0, 0, -20), timePtr(now.AddDate(0, 0, 10)), nil, "Active", nil)
	alloc5 := insAllocation(ctx, pool, officeChair, strPtr(frankID), strPtr(opsID), eveID, now.AddDate(0, 0, -15), nil, nil, "Active", nil)
	// A returned allocation
	insAllocation(ctx, pool, v1, nil, strPtr(mktID), carolID, now.AddDate(0, 0, -40), timePtr(now.AddDate(0, 0, -10)), timePtr(now.AddDate(0, 0, -12)), "Returned", nil)
	// An overdue allocation
	insAllocation(ctx, pool, soundSystem, strPtr(henryID), strPtr(itsID), eveID, now.AddDate(0, 0, -90), timePtr(now.AddDate(0, 0, -30)), nil, "Active", nil)
	// Another active
	insAllocation(ctx, pool, projector, nil, strPtr(hrID), graceID, now.AddDate(0, 0, -10), timePtr(now.AddDate(0, 0, 20)), nil, "Active", nil)

	// ===================================================================
	// TRANSFER REQUESTS (5 rows)
	// ===================================================================
	insTransfer(ctx, pool, laptop1, alloc1, strPtr(aliceID), strPtr(bobID), aliceID, "Requested")
	insTransfer(ctx, pool, laptop3, alloc3, strPtr(carolID), strPtr(davidID), carolID, "Requested")
	insTransfer(ctx, pool, monitor, alloc4, strPtr(davidID), strPtr(frankID), davidID, "Approved", strPtr(eveID))
	insTransfer(ctx, pool, officeChair, alloc5, strPtr(frankID), strPtr(henryID), frankID, "Rejected", strPtr(eveID))
	insTransfer(ctx, pool, laptop2, alloc2, strPtr(bobID), strPtr(aliceID), bobID, "Requested")

	// ===================================================================
	// BOOKINGS (12 rows — spread across days/hours for heatmap)
	// ===================================================================
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	for i := -5; i <= 6; i++ {
		day := today.AddDate(0, 0, i)
		switch i {
		case -5:
			insBooking(ctx, pool, rB1, aliceID, day.Add(9*time.Hour), day.Add(10*time.Hour), "Engineering sync")
			insBooking(ctx, pool, rB2, bobID, day.Add(14*time.Hour), day.Add(16*time.Hour), "Code review session")
		case -4:
			insBooking(ctx, pool, boardRoom, carolID, day.Add(10*time.Hour), day.Add(12*time.Hour), "Board meeting")
			insBooking(ctx, pool, trainingRoom, eveID, day.Add(8*time.Hour), day.Add(17*time.Hour), "All-hands training")
		case -3:
			insBooking(ctx, pool, v1, frankID, day.Add(9*time.Hour), day.Add(11*time.Hour), "Client site visit")
			insBooking(ctx, pool, rB1, graceID, day.Add(13*time.Hour), day.Add(14*time.Hour), "HR interview")
		case -2:
			insBooking(ctx, pool, v2, davidID, day.Add(10*time.Hour), day.Add(12*time.Hour), "Equipment pickup")
			insBooking(ctx, pool, projector, henryID, day.Add(15*time.Hour), day.Add(17*time.Hour), "Tech talk")
		case -1:
			insBooking(ctx, pool, rB2, aliceID, day.Add(9*time.Hour), day.Add(10*time.Hour), "Weekly sync")
		case 0:
			insBooking(ctx, pool, rB1, bobID, day.Add(13*time.Hour), day.Add(15*time.Hour), "Sprint planning")
			insBooking(ctx, pool, boardRoom, carolID, day.Add(11*time.Hour), day.Add(12*time.Hour), "Marketing review")
		case 1:
			insBooking(ctx, pool, v1, carolID, day.Add(10*time.Hour), day.Add(12*time.Hour), "Client visit")
			insBooking(ctx, pool, rB1, eveID, day.Add(14*time.Hour), day.Add(15*time.Hour), "Ops standup")
		case 2:
			insBooking(ctx, pool, trainingRoom, graceID, day.Add(9*time.Hour), day.Add(11*time.Hour), "New hire orientation")
		case 3:
			insBooking(ctx, pool, v2, frankID, day.Add(8*time.Hour), day.Add(10*time.Hour), "Field support")
			insBooking(ctx, pool, rB2, henryID, day.Add(15*time.Hour), day.Add(16*time.Hour), "IT training")
		case 4:
			insBooking(ctx, pool, projector, davidID, day.Add(13*time.Hour), day.Add(14*time.Hour), "Product demo")
		case 5:
			insBooking(ctx, pool, boardRoom, aliceID, day.Add(10*time.Hour), day.Add(12*time.Hour), "Quarterly review")
			insBooking(ctx, pool, v3, eveID, day.Add(14*time.Hour), day.Add(16*time.Hour), "Offsite errand")
		case 6:
			insBooking(ctx, pool, rB1, bobID, day.Add(9*time.Hour), day.Add(11*time.Hour), "Week planning")
		}
	}

	// ===================================================================
	// MAINTENANCE REQUESTS (8 rows)
	// ===================================================================
	insMaintenance(ctx, pool, projector, eveID, "Projector bulb flickering, needs replacement", "High", "TechnicianAssigned", strPtr("John Doe"), now.AddDate(0, 0, -5))
	insMaintenance(ctx, pool, soundSystem, henryID, "Left speaker produces crackling sound at high volume", "Medium", "Pending", nil, now.AddDate(0, 0, -3))
	insMaintenance(ctx, pool, v1, frankID, "Check engine light is on", "Critical", "InProgress", strPtr("Jane Smith"), now.AddDate(0, 0, -2))
	insMaintenance(ctx, pool, laptop1, bobID, "Battery drains quickly, lasts only 1 hour", "Medium", "Pending", nil, now.AddDate(0, 0, -1))
	insMaintenance(ctx, pool, laptop2, aliceID, "Keyboard keys E and R are stuck", "Low", "Resolved", strPtr("Mike Wilson"), now.AddDate(0, 0, -10))
	insMaintenance(ctx, pool, officeChair, frankID, "Hydraulic lift is not holding height", "Medium", "Approved", nil, now.AddDate(0, 0, -4))
	insMaintenance(ctx, pool, v2, davidID, "Air conditioning not cooling", "High", "TechnicianAssigned", strPtr("Jane Smith"), now.AddDate(0, 0, -6))
	insMaintenance(ctx, pool, monitor, davidID, "Dead pixel cluster in bottom-left corner", "Low", "Resolved", strPtr("Mike Wilson"), now.AddDate(0, 0, -20))

	// ===================================================================
	// AUDIT CYCLES (2 rows)
	// ===================================================================
	audit1ID := execScalar(ctx, pool,
		`INSERT INTO audit_cycles (name, scope_department_id, scope_location, start_date, end_date, status, created_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		"Q1 Engineering Asset Audit", engID, "10th Floor", today.AddDate(0, -2, 0), today.AddDate(0, -2, 14), "Closed", aliceID)

	audit2ID := execScalar(ctx, pool,
		`INSERT INTO audit_cycles (name, start_date, end_date, status, created_by)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		"Annual IT Equipment Review", today.AddDate(0, -1, 0), today.AddDate(0, 0, 30), "Active", eveID)

	// ===================================================================
	// AUDIT CYCLE AUDITORS (4 rows)
	// ===================================================================
	execVoid(ctx, pool, `INSERT INTO audit_cycle_auditors (audit_cycle_id, employee_id) VALUES ($1, $2)`, audit1ID, aliceID)
	execVoid(ctx, pool, `INSERT INTO audit_cycle_auditors (audit_cycle_id, employee_id) VALUES ($1, $2)`, audit1ID, bobID)
	execVoid(ctx, pool, `INSERT INTO audit_cycle_auditors (audit_cycle_id, employee_id) VALUES ($1, $2)`, audit2ID, eveID)
	execVoid(ctx, pool, `INSERT INTO audit_cycle_auditors (audit_cycle_id, employee_id) VALUES ($1, $2)`, audit2ID, henryID)

	// ===================================================================
	// AUDIT ITEMS (8 rows)
	// ===================================================================
	item1ID := insAuditItem(ctx, pool, audit1ID, laptop1, aliceID, "Verified", "In good condition", now.AddDate(0, -2, 5))
	item2ID := insAuditItem(ctx, pool, audit1ID, laptop2, aliceID, "Verified", "Minor wear on keyboard", now.AddDate(0, -2, 5))
	insAuditItem(ctx, pool, audit1ID, monitor, aliceID, "Damaged", "Screen has a small crack", now.AddDate(0, -2, 6))
	insAuditItem(ctx, pool, audit1ID, officeChair, bobID, "Verified", "Good condition", now.AddDate(0, -2, 6))
	insAuditItem(ctx, pool, audit2ID, server, eveID, "Verified", "Running smoothly", now.AddDate(0, -1, 3))
	insAuditItem(ctx, pool, audit2ID, laptop3, eveID, "Missing", "Could not locate the asset", now.AddDate(0, -1, 3))
	insAuditItem(ctx, pool, audit2ID, laptop1, henryID, "Verified", "OK", now.AddDate(0, -1, 4))
	insAuditItem(ctx, pool, audit2ID, monitor, henryID, "Verified", "Screen replaced", now.AddDate(0, -1, 4))

	// ===================================================================
	// DISCREPANCY REPORTS (3 rows)
	// ===================================================================
	execVoid(ctx, pool,
		`INSERT INTO discrepancy_reports (audit_cycle_id, asset_id, audit_item_id, issue_type)
		 VALUES ($1, $2, $3, 'Damaged')`, audit1ID, monitor, item1ID)
	execVoid(ctx, pool,
		`INSERT INTO discrepancy_reports (audit_cycle_id, asset_id, audit_item_id, issue_type, resolved, resolved_by, resolved_at)
		 VALUES ($1, $2, $3, 'Damaged', true, $4, $5)`, audit1ID, monitor, item2ID, eveID, now.AddDate(0, -2, 10))
	execVoid(ctx, pool,
		`INSERT INTO discrepancy_reports (audit_cycle_id, asset_id, audit_item_id, issue_type)
		 VALUES ($1, $2, $3, 'Missing')`, audit2ID, laptop3, item2ID)

	// ===================================================================
	// NOTIFICATIONS (10 rows)
	// ===================================================================
	insNotification(ctx, pool, aliceID, "AssetAssigned", "Dell Latitude 5540 has been assigned to you", "allocation", strPtr(alloc1))
	insNotification(ctx, pool, bobID, "AssetAssigned", "MacBook Pro 16 has been assigned to you", "allocation", strPtr(alloc2))
	insNotification(ctx, pool, carolID, "AssetAssigned", "Lenovo ThinkPad X1 has been assigned to you", "allocation", strPtr(alloc3))
	insNotification(ctx, pool, frankID, "MaintenanceResolved", "Your chair maintenance request has been resolved", "maintenance_request", strPtr(alloc5))
	insNotification(ctx, pool, aliceID, "BookingReminder", "Conference Room B booking starts in 30 min", "booking", nil)
	insNotification(ctx, pool, bobID, "BookingReminder", "Conference Room A booking starts in 1 hour", "booking", nil)
	insNotification(ctx, pool, eveID, "AuditDiscrepancyFlagged", "Missing asset found in audit cycle", "audit_cycle", strPtr(audit2ID))
	insNotification(ctx, pool, henryID, "MaintenanceAssigned", "Sound System maintenance ticket assigned to you", "maintenance_request", nil)
	insNotification(ctx, pool, graceID, "TransferRequest", "Transfer request for office chair needs your review", "transfer", nil)
	insNotification(ctx, pool, davidID, "ReturnOverdue", "Monitor allocation is overdue", "allocation", strPtr(alloc4))

	// ===================================================================
	// ACTIVITY LOGS (15 rows)
	// ===================================================================
	insActivityLog(ctx, pool, aliceID, "asset.create", "asset", strPtr(laptop1), now.AddDate(0, 0, -60))
	insActivityLog(ctx, pool, aliceID, "asset.allocate", "allocation", strPtr(alloc1), now.AddDate(0, 0, -60))
	insActivityLog(ctx, pool, eveID, "asset.create", "asset", strPtr(server), now.AddDate(0, 0, -50))
	insActivityLog(ctx, pool, carolID, "asset.allocate", "allocation", strPtr(alloc3), now.AddDate(0, 0, -30))
	insActivityLog(ctx, pool, eveID, "maintenance.create", "maintenance_request", nil, now.AddDate(0, 0, -5))
	insActivityLog(ctx, pool, eveID, "maintenance.resolve", "maintenance_request", nil, now.AddDate(0, 0, -4))
	insActivityLog(ctx, pool, frankID, "booking.create", "booking", nil, now.AddDate(0, 0, -3))
	insActivityLog(ctx, pool, aliceID, "booking.create", "booking", nil, now.AddDate(0, 0, -2))
	insActivityLog(ctx, pool, bobID, "booking.cancel", "booking", nil, now.AddDate(0, 0, -2))
	insActivityLog(ctx, pool, carolID, "transfer.request", "transfer", nil, now.AddDate(0, 0, -1))
	insActivityLog(ctx, pool, eveID, "transfer.approve", "transfer", nil, now.AddDate(0, 0, -1))
	insActivityLog(ctx, pool, eveID, "audit.create", "audit_cycle", strPtr(audit2ID), now.AddDate(0, -1, 0))
	insActivityLog(ctx, pool, eveID, "audit.close", "audit_cycle", strPtr(audit1ID), now.AddDate(0, -2, 15))
	insActivityLog(ctx, pool, graceID, "employee.create", "employee", nil, now.AddDate(0, 0, -20))
	insActivityLog(ctx, pool, aliceID, "department.update", "department", strPtr(engID), now.AddDate(0, 0, -15))

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

// ── helpers ─────────────────────────────────────────────────────────────

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
		`INSERT INTO assets (name, category_id, is_bookable, status, custom_fields, acquisition_date)
		 VALUES ($1, $2, true, 'Available', $3::jsonb, $4) RETURNING id`,
		name, categoryID, customFields, timePtr(time.Now().AddDate(-3, 0, 0)))
}

func insNonBookableGetID(ctx context.Context, pool *pgxpool.Pool, name, categoryID, serial string, employeeID *string, status string, acquisitionDate *time.Time) string {
	if employeeID != nil {
		return execScalar(ctx, pool,
			`INSERT INTO assets (name, category_id, is_bookable, status, serial_number, current_holder_employee_id, acquisition_date)
			 VALUES ($1, $2, false, $3, $4, $5, $6) RETURNING id`,
			name, categoryID, status, serial, *employeeID, acquisitionDate)
	}
	return execScalar(ctx, pool,
		`INSERT INTO assets (name, category_id, is_bookable, status, serial_number, acquisition_date)
		 VALUES ($1, $2, false, $3, $4, $5) RETURNING id`,
		name, categoryID, status, serial, acquisitionDate)
}

func insStatusHistory(ctx context.Context, pool *pgxpool.Pool, assetID string, fromStatus, toStatus *string, changedBy string, reason string) {
	execVoid(ctx, pool,
		`INSERT INTO asset_status_history (asset_id, from_status, to_status, changed_by, reason)
		 VALUES ($1, $2, $3, $4, $5)`,
		assetID, fromStatus, toStatus, changedBy, reason)
}

func insAllocation(ctx context.Context, pool *pgxpool.Pool, assetID string, employeeID *string, departmentID *string, allocatedBy string, allocatedAt time.Time, expectedReturnDate *time.Time, returnedAt *time.Time, status string, returnConditionNotes *string) string {
	return execScalar(ctx, pool,
		`INSERT INTO allocations (asset_id, employee_id, department_id, allocated_by, allocated_at, expected_return_date, returned_at, status, return_condition_notes)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`,
		assetID, employeeID, departmentID, allocatedBy, allocatedAt, expectedReturnDate, returnedAt, status, returnConditionNotes)
}

func insTransfer(ctx context.Context, pool *pgxpool.Pool, assetID, allocationID string, fromEmployeeID, toEmployeeID *string, requestedBy string, status string, approvedBy ...*string) string {
	var appBy *string
	if len(approvedBy) > 0 {
		appBy = approvedBy[0]
	}
	return execScalar(ctx, pool,
		`INSERT INTO transfer_requests (asset_id, allocation_id, from_employee_id, to_employee_id, requested_by, status, approved_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		assetID, allocationID, fromEmployeeID, toEmployeeID, requestedBy, status, appBy)
}

func insBooking(ctx context.Context, pool *pgxpool.Pool, resourceID, employeeID string, start, end time.Time, purpose string) {
	statuses := []string{"Completed", "Completed", "Completed", "Completed", "Completed", "Completed", "Completed", "Upcoming", "Upcoming", "Upcoming", "Upcoming", "Ongoing"}
	idx := rng.Intn(len(statuses))
	execScalar(ctx, pool,
		`INSERT INTO bookings (resource_asset_id, booked_by_employee_id, start_time, end_time, purpose, status, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		resourceID, employeeID, start, end, purpose, statuses[idx], start.Add(-1*time.Hour))
}

func insMaintenance(ctx context.Context, pool *pgxpool.Pool, assetID, raisedByID, description, priority, status string, technician *string, createdAt time.Time) {
	execScalar(ctx, pool,
		`INSERT INTO maintenance_requests (asset_id, raised_by_employee_id, issue_description, priority, status, technician_name, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		assetID, raisedByID, description, priority, status, technician, createdAt)
}

func insAuditItem(ctx context.Context, pool *pgxpool.Pool, cycleID, assetID, auditorID, result, notes string, verifiedAt time.Time) string {
	return execScalar(ctx, pool,
		`INSERT INTO audit_items (audit_cycle_id, asset_id, auditor_id, result, notes, verified_at, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		cycleID, assetID, auditorID, result, notes, verifiedAt, verifiedAt)
}

func insNotification(ctx context.Context, pool *pgxpool.Pool, employeeID, notifType, message, relatedEntityType string, relatedEntityID *string) {
	execVoid(ctx, pool,
		`INSERT INTO notifications (employee_id, type, message, related_entity_type, related_entity_id)
		 VALUES ($1, $2, $3, $4, $5)`,
		employeeID, notifType, message, relatedEntityType, relatedEntityID)
}

func insActivityLog(ctx context.Context, pool *pgxpool.Pool, actorID, action, entityType string, entityID *string, createdAt time.Time) {
	execVoid(ctx, pool,
		`INSERT INTO activity_logs (actor_employee_id, action, entity_type, entity_id, created_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		actorID, action, entityType, entityID, createdAt)
}

func strPtr(s string) *string { return &s }

func timePtr(t time.Time) *time.Time { return &t }
