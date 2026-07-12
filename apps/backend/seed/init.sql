CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "btree_gist";

CREATE TYPE role_enum AS ENUM ('Admin', 'DepartmentHead', 'AssetManager', 'Employee');
CREATE TYPE status_enum AS ENUM ('Active', 'Inactive');

CREATE TYPE asset_status_enum AS ENUM (
    'Available', 'Allocated', 'Reserved', 'UnderMaintenance', 'Lost', 'Retired', 'Disposed'
);

CREATE TYPE allocation_status_enum AS ENUM ('Active', 'Returned', 'Overdue');
CREATE TYPE transfer_status_enum AS ENUM ('Requested', 'Approved', 'Rejected');
CREATE TYPE booking_status_enum AS ENUM ('Upcoming', 'Ongoing', 'Completed', 'Cancelled');

CREATE TYPE maintenance_status_enum AS ENUM (
    'Pending', 'Approved', 'Rejected', 'TechnicianAssigned', 'InProgress', 'Resolved'
);
CREATE TYPE maintenance_priority_enum AS ENUM ('Low', 'Medium', 'High', 'Critical');

CREATE TYPE audit_cycle_status_enum AS ENUM ('Draft', 'Active', 'Closed');
CREATE TYPE audit_item_result_enum AS ENUM ('Verified', 'Missing', 'Damaged');

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- =====================================================================
-- MODULE: Organization Setup (Departments, Categories, Employees)
-- =====================================================================

-- Departments (self-referencing hierarchy; head_employee_id FK added after employees exists)
CREATE TABLE departments (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name                 VARCHAR(150) NOT NULL UNIQUE,
    parent_department_id UUID REFERENCES departments(id) ON DELETE SET NULL,
    head_employee_id     UUID, -- FK added below, after employees table is created
    status               status_enum NOT NULL DEFAULT 'Active',
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TRIGGER trg_departments_updated_at
    BEFORE UPDATE ON departments
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Asset categories (custom fields per category, e.g. warranty_period for Electronics)
CREATE TABLE asset_categories (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name          VARCHAR(150) NOT NULL UNIQUE,
    custom_fields JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TRIGGER trg_categories_updated_at
    BEFORE UPDATE ON asset_categories
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Employees (the directory — every user is an employee first; role assigned only via Admin promotion)
CREATE TABLE employees (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name          VARCHAR(150) NOT NULL,
    email         VARCHAR(255) NOT NULL UNIQUE,
    department_id UUID REFERENCES departments(id) ON DELETE SET NULL,
    role          role_enum NOT NULL DEFAULT 'Employee',
    status        status_enum NOT NULL DEFAULT 'Active',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TRIGGER trg_employees_updated_at
    BEFORE UPDATE ON employees
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Now that employees exists, wire up the department head FK
ALTER TABLE departments
    ADD CONSTRAINT fk_departments_head_employee
    FOREIGN KEY (head_employee_id) REFERENCES employees(id) ON DELETE SET NULL;

CREATE INDEX idx_employees_department ON employees(department_id);
CREATE INDEX idx_employees_role ON employees(role);

-- Auth: one users row per employee (keeps auth concerns separate from directory data)
CREATE TABLE users (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id        UUID NOT NULL UNIQUE REFERENCES employees(id) ON DELETE CASCADE,
    password_hash      VARCHAR(255) NOT NULL,
    refresh_token_hash VARCHAR(255),
    last_login_at      TIMESTAMPTZ,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TRIGGER trg_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- =====================================================================
-- MODULE: Asset Registration & Directory
-- =====================================================================

-- Sequence + helper to auto-generate asset tags like AF-0001, AF-0002, ...
CREATE SEQUENCE asset_tag_seq START 1;

CREATE OR REPLACE FUNCTION generate_asset_tag()
RETURNS TEXT AS $$
BEGIN
    RETURN 'AF-' || LPAD(nextval('asset_tag_seq')::TEXT, 4, '0');
END;
$$ LANGUAGE plpgsql;

CREATE TABLE assets (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    asset_tag                   VARCHAR(20) NOT NULL UNIQUE DEFAULT generate_asset_tag(),
    name                        VARCHAR(200) NOT NULL,
    category_id                 UUID NOT NULL REFERENCES asset_categories(id),
    serial_number               VARCHAR(150),
    acquisition_date            DATE,
    acquisition_cost            NUMERIC(14, 2),
    condition                   VARCHAR(50),
    location                    VARCHAR(200),
    is_bookable                 BOOLEAN NOT NULL DEFAULT false,
    status                      asset_status_enum NOT NULL DEFAULT 'Available',
    current_holder_employee_id  UUID REFERENCES employees(id) ON DELETE SET NULL,
    current_holder_department_id UUID REFERENCES departments(id) ON DELETE SET NULL,
    qr_code                     VARCHAR(255),
    custom_fields               JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TRIGGER trg_assets_updated_at
    BEFORE UPDATE ON assets
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE INDEX idx_assets_status ON assets(status);
CREATE INDEX idx_assets_category ON assets(category_id);
CREATE INDEX idx_assets_location ON assets(location);
CREATE INDEX idx_assets_serial ON assets(serial_number);
CREATE INDEX idx_assets_bookable ON assets(is_bookable) WHERE is_bookable = true;

-- Photos / documents attached to an asset
CREATE TABLE asset_documents (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    asset_id    UUID NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    url         VARCHAR(500) NOT NULL,
    type        VARCHAR(20) NOT NULL CHECK (type IN ('photo', 'document')),
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_asset_documents_asset ON asset_documents(asset_id);

-- Full lifecycle audit trail — every status change writes a row here
CREATE TABLE asset_status_history (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    asset_id    UUID NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    from_status asset_status_enum,
    to_status   asset_status_enum NOT NULL,
    changed_by  UUID REFERENCES employees(id),
    reason      VARCHAR(255),
    changed_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_asset_status_history_asset ON asset_status_history(asset_id, changed_at DESC);

-- =====================================================================
-- MODULE: Allocation & Transfer
-- =====================================================================

CREATE TABLE allocations (
    id                     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    asset_id               UUID NOT NULL REFERENCES assets(id),
    employee_id            UUID REFERENCES employees(id),
    department_id          UUID REFERENCES departments(id),
    allocated_by           UUID NOT NULL REFERENCES employees(id),
    allocated_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    expected_return_date   DATE,
    returned_at            TIMESTAMPTZ,
    return_condition_notes TEXT,
    status                 allocation_status_enum NOT NULL DEFAULT 'Active',
    created_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_allocation_holder CHECK (employee_id IS NOT NULL OR department_id IS NOT NULL)
);

CREATE TRIGGER trg_allocations_updated_at
    BEFORE UPDATE ON allocations
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- The core "no double allocation" rule, enforced at the DB level:
-- only one Active allocation may exist per asset at any time.
CREATE UNIQUE INDEX ux_allocations_one_active_per_asset
    ON allocations(asset_id) WHERE status = 'Active';

CREATE INDEX idx_allocations_employee ON allocations(employee_id);
CREATE INDEX idx_allocations_department ON allocations(department_id);
CREATE INDEX idx_allocations_overdue ON allocations(expected_return_date)
    WHERE status = 'Active';

CREATE TABLE transfer_requests (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    asset_id         UUID NOT NULL REFERENCES assets(id),
    allocation_id    UUID NOT NULL REFERENCES allocations(id),
    from_employee_id UUID REFERENCES employees(id),
    to_employee_id   UUID NOT NULL REFERENCES employees(id),
    requested_by     UUID NOT NULL REFERENCES employees(id),
    status           transfer_status_enum NOT NULL DEFAULT 'Requested',
    approved_by      UUID REFERENCES employees(id),
    approved_at      TIMESTAMPTZ,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TRIGGER trg_transfer_requests_updated_at
    BEFORE UPDATE ON transfer_requests
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE INDEX idx_transfer_requests_status ON transfer_requests(status);
CREATE INDEX idx_transfer_requests_asset ON transfer_requests(asset_id);

-- =====================================================================
-- MODULE: Resource Booking
-- (bookable "resources" are just assets where is_bookable = true)
-- =====================================================================

CREATE TABLE bookings (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    resource_asset_id    UUID NOT NULL REFERENCES assets(id),
    booked_by_employee_id UUID NOT NULL REFERENCES employees(id),
    start_time           TIMESTAMPTZ NOT NULL,
    end_time             TIMESTAMPTZ NOT NULL,
    purpose              VARCHAR(255),
    status               booking_status_enum NOT NULL DEFAULT 'Upcoming',
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_booking_time_order CHECK (end_time > start_time)
);

CREATE TRIGGER trg_bookings_updated_at
    BEFORE UPDATE ON bookings
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- The core "no overlapping bookings" rule, enforced at the DB level.
-- A request for 9:30-10:30 against an existing 9:00-10:00 booking will
-- fail here even if the application layer has a bug; a 10:00-11:00
-- request succeeds since ranges are half-open (touching, not overlapping).
ALTER TABLE bookings
    ADD CONSTRAINT excl_bookings_no_overlap
    EXCLUDE USING gist (
        resource_asset_id WITH =,
        tstzrange(start_time, end_time) WITH &&
    ) WHERE (status <> 'Cancelled');

CREATE INDEX idx_bookings_resource_time ON bookings(resource_asset_id, start_time, end_time);
CREATE INDEX idx_bookings_status ON bookings(status);

-- =====================================================================
-- MODULE: Maintenance Management
-- =====================================================================

CREATE TABLE maintenance_requests (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    asset_id            UUID NOT NULL REFERENCES assets(id),
    raised_by_employee_id UUID NOT NULL REFERENCES employees(id),
    issue_description   TEXT NOT NULL,
    priority             maintenance_priority_enum NOT NULL DEFAULT 'Medium',
    photo_url            VARCHAR(500),
    status               maintenance_status_enum NOT NULL DEFAULT 'Pending',
    approved_by          UUID REFERENCES employees(id),
    approved_at          TIMESTAMPTZ,
    technician_name      VARCHAR(150),
    resolved_at          TIMESTAMPTZ,
    resolution_notes     TEXT,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TRIGGER trg_maintenance_requests_updated_at
    BEFORE UPDATE ON maintenance_requests
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE INDEX idx_maintenance_asset ON maintenance_requests(asset_id);
CREATE INDEX idx_maintenance_status ON maintenance_requests(status);

-- =====================================================================
-- MODULE: Asset Audit
-- =====================================================================

CREATE TABLE audit_cycles (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name                 VARCHAR(200) NOT NULL,
    scope_department_id  UUID REFERENCES departments(id),
    scope_location       VARCHAR(200),
    start_date           DATE NOT NULL,
    end_date             DATE NOT NULL,
    status               audit_cycle_status_enum NOT NULL DEFAULT 'Draft',
    created_by           UUID NOT NULL REFERENCES employees(id),
    closed_at            TIMESTAMPTZ,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_audit_cycle_dates CHECK (end_date >= start_date)
);

CREATE TRIGGER trg_audit_cycles_updated_at
    BEFORE UPDATE ON audit_cycles
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE audit_cycle_auditors (
    audit_cycle_id UUID NOT NULL REFERENCES audit_cycles(id) ON DELETE CASCADE,
    employee_id    UUID NOT NULL REFERENCES employees(id),
    PRIMARY KEY (audit_cycle_id, employee_id)
);

CREATE TABLE audit_items (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    audit_cycle_id UUID NOT NULL REFERENCES audit_cycles(id) ON DELETE CASCADE,
    asset_id       UUID NOT NULL REFERENCES assets(id),
    auditor_id     UUID REFERENCES employees(id),
    result         audit_item_result_enum,
    notes          TEXT,
    verified_at    TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (audit_cycle_id, asset_id)
);

CREATE TRIGGER trg_audit_items_updated_at
    BEFORE UPDATE ON audit_items
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE INDEX idx_audit_items_cycle ON audit_items(audit_cycle_id);
CREATE INDEX idx_audit_items_result ON audit_items(result);

-- Auto-generated whenever an audit_item is marked Missing or Damaged
CREATE TABLE discrepancy_reports (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    audit_cycle_id UUID NOT NULL REFERENCES audit_cycles(id) ON DELETE CASCADE,
    asset_id       UUID NOT NULL REFERENCES assets(id),
    audit_item_id  UUID NOT NULL REFERENCES audit_items(id),
    issue_type     VARCHAR(20) NOT NULL CHECK (issue_type IN ('Missing', 'Damaged')),
    resolved       BOOLEAN NOT NULL DEFAULT false,
    resolved_by    UUID REFERENCES employees(id),
    resolved_at    TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TRIGGER trg_discrepancy_reports_updated_at
    BEFORE UPDATE ON discrepancy_reports
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE INDEX idx_discrepancy_reports_cycle ON discrepancy_reports(audit_cycle_id);
CREATE INDEX idx_discrepancy_reports_resolved ON discrepancy_reports(resolved);

-- =====================================================================
-- MODULE: Notifications & Activity Logs (cross-cutting)
-- =====================================================================

CREATE TABLE notifications (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id          UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    type                 VARCHAR(50) NOT NULL, -- e.g. 'AssetAssigned', 'BookingReminder', 'AuditDiscrepancyFlagged'
    message              TEXT NOT NULL,
    related_entity_type  VARCHAR(50),
    related_entity_id    UUID,
    is_read              BOOLEAN NOT NULL DEFAULT false,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_notifications_employee_unread ON notifications(employee_id, is_read);
CREATE INDEX idx_notifications_created ON notifications(created_at DESC);

CREATE TABLE activity_logs (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_employee_id UUID REFERENCES employees(id),
    action            VARCHAR(100) NOT NULL, -- e.g. 'asset.allocate', 'booking.cancel', 'audit.close'
    entity_type       VARCHAR(50) NOT NULL,
    entity_id         UUID,
    metadata          JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_activity_logs_entity ON activity_logs(entity_type, entity_id);
CREATE INDEX idx_activity_logs_actor ON activity_logs(actor_employee_id);
CREATE INDEX idx_activity_logs_created ON activity_logs(created_at DESC);

-- Default admin is seeded programmatically in init.go with a real bcrypt hash.
