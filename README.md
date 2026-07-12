🏢 AssetHub — Enterprise Asset Management System

> A full-stack enterprise asset management platform built for the **Odoo Hackathon 2026**. Track, allocate, maintain, audit, and report on organizational assets with role-based access control, real-time notifications, and rich analytics.

---

## 📑 Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Tech Stack](#tech-stack)
- [Architecture](#architecture)
  - [High-Level Architecture](#high-level-architecture)
  - [Project Structure](#project-structure)
- [Data Model](#data-model)
  - [Entity Relationship Diagram](#entity-relationship-diagram)
  - [Entity Descriptions](#entity-descriptions)
- [Authentication & Authorization](#authentication--authorization)
  - [Auth Flow](#auth-flow)
  - [Role-Based Access Control](#role-based-access-control)
- [API Reference](#api-reference)
  - [Auth Endpoints](#auth-endpoints)
  - [Asset Endpoints](#asset-endpoints)
  - [Allocation & Transfer Endpoints](#allocation--transfer-endpoints)
  - [Maintenance Endpoints](#maintenance-endpoints)
  - [Booking Endpoints](#booking-endpoints)
  - [Audit Endpoints](#audit-endpoints)
  - [Dashboard & Reports](#dashboard--reports-endpoints)
  - [Notifications & Activity Logs](#notifications--activity-log-endpoints)
  - [Admin Endpoints](#admin-endpoints)
- [Business Workflows](#business-workflows)
  - [Asset Lifecycle](#asset-lifecycle)
  - [Allocation & Transfer Flow](#allocation--transfer-workflow)
  - [Maintenance Workflow](#maintenance-workflow)
  - [Audit Workflow](#audit-workflow)
  - [Booking Workflow](#booking-workflow)
- [Background Jobs](#background-jobs)
- [Frontend Pages](#frontend-pages)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Local Development](#local-development)
  - [Docker Development](#docker-development)
- [Environment Variables](#environment-variables)
- [Task Runner Commands](#task-runner-commands)

---

## Overview

**AssetHub** is a comprehensive enterprise asset management system that enables organizations to:

- **Register & catalog** all physical and digital assets with categories, serial numbers, purchase info, and warranty tracking
- **Allocate assets** to employees with due-date tracking and automated overdue detection
- **Transfer assets** between employees via an approval-based workflow
- **Schedule & track maintenance** through a multi-stage approval pipeline
- **Book shared resources** (conference rooms, projectors, vehicles) with conflict detection
- **Conduct periodic audits** with auditor assignment, item verification, and discrepancy reporting
- **Generate analytics & reports** — utilization rates, maintenance frequency, retirement watchlists, booking heatmaps, and CSV exports
- **Receive real-time notifications** for approvals, overdue items, and system events

---

## Features

| Category | Capabilities |
|---|---|
| **Asset Management** | Register, search, filter, view detailed history, upload documents |
| **Allocation** | Assign assets to employees, track expected returns, auto-detect overdue |
| **Transfers** | Employee-initiated transfer requests with manager approval |
| **Maintenance** | Multi-stage workflow: Request → Approve → Assign Technician → Start → Resolve |
| **Resource Booking** | Time-slot booking with conflict detection, cancel/reschedule support |
| **Auditing** | Create audit cycles, assign auditors, verify assets, report discrepancies |
| **Dashboard** | KPI cards, overdue alerts, upcoming maintenance, recent activity feed |
| **Reports** | Utilization, maintenance frequency, retirement watchlist, booking heatmaps, CSV export |
| **Notifications** | In-app notifications with unread counts and bulk mark-as-read |
| **RBAC** | Four-tier role system: Admin, AssetManager, DepartmentHead, Employee |

---

## Tech Stack

### Backend

| Technology | Purpose |
|---|---|
| **Go 1.25** | Server language |
| **Gin** | HTTP web framework |
| **PostgreSQL 15** | Relational database |
| **pgx/v5** | PostgreSQL driver & connection pooling |
| **golang-jwt/v5** | JWT authentication |
| **bcrypt** | Password hashing |
| **Air** | Hot-reload for development |

### Frontend

| Technology | Purpose |
|---|---|
| **React 19** | UI framework |
| **TypeScript 6** | Type-safe JavaScript |
| **Vite 8** | Build tool & dev server |
| **TanStack Router** | File-based type-safe routing |
| **TanStack React Query** | Server state management & caching |
| **Zustand** | Client state management (auth) |
| **Axios** | HTTP client with interceptors |
| **Tailwind CSS 4** | Utility-first styling |
| **shadcn/ui** | Component library (Base UI) |
| **Recharts 3** | Data visualization & charts |
| **Lucide React** | Icon library |

### Infrastructure

| Technology | Purpose |
|---|---|
| **Docker & Docker Compose** | Containerized development |
| **Task (Taskfile)** | Task runner for development workflows |
| **pnpm** | Frontend package manager |

---

## Architecture

### High-Level Architecture

```mermaid
graph TB
    subgraph Client["🌐 Browser"]
        FE["React SPA<br/>(Vite + TanStack Router)"]
    end

    subgraph Docker["🐳 Docker Compose"]
        subgraph Backend["Go Backend"]
            GIN["Gin Router<br/>:8000"]
            MW["Middleware<br/>(Auth + RBAC)"]
            H["Handlers"]
            R["Repository Layer"]
            SCH["Scheduler<br/>(Background Jobs)"]
        end

        subgraph Database["PostgreSQL 15"]
            DB[("odoo-hackathon<br/>:5432")]
        end

        subgraph Frontend["Frontend Dev Server"]
            VITE["Vite Dev Server<br/>:5173"]
        end
    end

    FE -->|"HTTP REST<br/>JSON + JWT"| GIN
    GIN --> MW --> H --> R --> DB
    SCH -->|"Periodic Tasks"| R
    VITE -->|"Serves"| FE

    style Client fill:#1e293b,stroke:#3b82f6,color:#e2e8f0
    style Docker fill:#0f172a,stroke:#6366f1,color:#e2e8f0
    style Backend fill:#1e1b4b,stroke:#818cf8,color:#e2e8f0
    style Database fill:#1e3a2f,stroke:#4ade80,color:#e2e8f0
    style Frontend fill:#1e293b,stroke:#f472b6,color:#e2e8f0
```

### Project Structure

```
odoo-hackathon-2026/
├── docker-compose.yml          # Multi-service container orchestration
├── Taskfile.yml                # Task runner definitions
├── .env.example                # Environment variable template
│
├── apps/
│   ├── backend/                # Go REST API
│   │   ├── main.go             # Entry point, router setup, DI
│   │   ├── Dockerfile          # Multi-stage Docker build
│   │   ├── .air.toml           # Hot-reload configuration
│   │   ├── go.mod / go.sum     # Go module dependencies
│   │   ├── seed/               # Database migrations & seed data
│   │   │   └── seed.go
│   │   └── internals/
│   │       ├── auth/           # JWT token generation & validation
│   │       ├── handlers/       # HTTP request handlers (by domain)
│   │       │   ├── activitylog/
│   │       │   ├── allocation/
│   │       │   ├── asset/
│   │       │   ├── audit/
│   │       │   ├── auth/
│   │       │   ├── booking/
│   │       │   ├── category/
│   │       │   ├── dashboard/
│   │       │   ├── department/
│   │       │   ├── employee/
│   │       │   ├── maintenance/
│   │       │   ├── notification/
│   │       │   └── report/
│   │       ├── middleware/     # Auth & RBAC middleware
│   │       ├── models/        # Domain entity structs
│   │       ├── repository/    # Database access layer (pgx)
│   │       └── scheduler/     # Background job runners
│   │
│   └── frontend/              # React SPA
│       ├── Dockerfile          # Multi-stage Docker build
│       ├── index.html          # SPA entry HTML
│       ├── package.json        # Dependencies & scripts
│       ├── vite.config.ts      # Vite configuration
│       ├── tsr.config.json     # TanStack Router config
│       ├── components.json     # shadcn/ui configuration
│       └── src/
│           ├── main.tsx        # React entry point
│           ├── router.tsx      # Router instance creation
│           ├── index.css       # Global styles & Tailwind
│           ├── components/     # Reusable UI components
│           │   ├── app-layout.tsx
│           │   ├── app-sidebar.tsx
│           │   └── ui/        # shadcn/ui primitives
│           ├── hooks/         # Custom React hooks
│           │   ├── use-auth.ts
│           │   ├── use-mobile.tsx
│           │   └── use-toast.ts
│           ├── lib/           # Utilities & API client
│           │   ├── api.ts
│           │   └── utils.ts
│           └── routes/        # File-based route pages
│               ├── __root.tsx
│               ├── login.tsx
│               ├── signup.tsx
│               └── _app/     # Authenticated routes
│                   ├── index.tsx         # Dashboard
│                   ├── assets/
│                   ├── allocations/
│                   ├── transfers/
│                   ├── maintenance/
│                   ├── bookings/
│                   ├── audit/
│                   ├── reports/
│                   ├── departments/
│                   ├── employees/
│                   ├── notifications/
│                   └── settings/
```

### Request Lifecycle

```mermaid
sequenceDiagram
    participant B as Browser
    participant G as Gin Router
    participant A as AuthMiddleware
    participant R as RoleMiddleware
    participant H as Handler
    participant S as Repository
    participant D as PostgreSQL

    B->>G: HTTP Request + JWT
    G->>A: Route matched
    A->>A: Validate JWT token
    alt Invalid/Expired Token
        A-->>B: 401 Unauthorized
    end
    A->>R: Set user context
    R->>R: Check role permissions
    alt Insufficient Role
        R-->>B: 403 Forbidden
    end
    R->>H: Authorized request
    H->>H: Validate input / business logic
    H->>S: Database operation
    S->>D: SQL query (pgx)
    D-->>S: Result rows
    S-->>H: Typed Go structs
    H-->>B: JSON response
```

---

## Data Model

### Entity Relationship Diagram

```mermaid
erDiagram
    Department ||--o{ Employee : "has many"
    Department ||--o{ Asset : "owns"
    Category ||--o{ Asset : "classifies"
    Employee ||--o{ Asset : "currently holds"
    
    Asset ||--o{ Allocation : "allocated via"
    Employee ||--o{ Allocation : "receives"
    Employee ||--o{ Allocation : "allocated by"
    
    Allocation ||--o{ TransferRequest : "can transfer"
    Employee ||--o{ TransferRequest : "from"
    Employee ||--o{ TransferRequest : "to"
    
    Asset ||--o{ MaintenanceRequest : "maintained"
    Employee ||--o{ MaintenanceRequest : "requested by"
    Employee ||--o{ MaintenanceRequest : "assigned to"
    
    Asset ||--o{ Booking : "booked"
    Employee ||--o{ Booking : "booked by"
    
    AuditCycle ||--o{ AuditItem : "contains"
    Asset ||--o{ AuditItem : "audited"
    Employee ||--o{ AuditItem : "verified by"
    
    AuditItem ||--o| DiscrepancyReport : "may report"
    Employee ||--o{ DiscrepancyReport : "reported by"
    Employee ||--o{ DiscrepancyReport : "resolved by"
    
    Employee ||--o{ Notification : "receives"
    Employee ||--o{ ActivityLog : "performed by"
    Asset ||--o{ AssetDocument : "has"

    Department {
        uuid id PK
        string name
        string description
        bool is_active
        timestamp created_at
        timestamp updated_at
    }

    Employee {
        uuid id PK
        string name
        string email UK
        string password_hash
        enum role "Admin | AssetManager | DepartmentHead | Employee"
        uuid department_id FK
        bool is_active
        timestamp created_at
        timestamp updated_at
    }

    Category {
        uuid id PK
        string name
        string description
        timestamp created_at
        timestamp updated_at
    }

    Asset {
        uuid id PK
        string name
        string serial_number UK
        uuid category_id FK
        uuid department_id FK
        enum status "Available | Allocated | UnderMaintenance | Retired | Lost"
        date purchase_date
        decimal purchase_price
        date warranty_expiry
        text notes
        uuid current_employee_id FK
        timestamp created_at
        timestamp updated_at
    }

    Allocation {
        uuid id PK
        uuid asset_id FK
        uuid employee_id FK
        uuid allocated_by FK
        timestamp allocated_at
        timestamp expected_return
        timestamp returned_at
        text notes
        enum status "Active | Returned | Overdue | TransferPending"
    }

    TransferRequest {
        uuid id PK
        uuid allocation_id FK
        uuid from_employee_id FK
        uuid to_employee_id FK
        text reason
        enum status "Pending | Approved | Rejected"
        uuid approved_by FK
        timestamp created_at
        timestamp updated_at
    }

    MaintenanceRequest {
        uuid id PK
        uuid asset_id FK
        uuid requested_by FK
        text description
        enum priority "Low | Medium | High | Critical"
        enum status "Pending | Approved | Rejected | InProgress | Resolved"
        uuid assigned_to FK
        text resolution_notes
        timestamp created_at
        timestamp updated_at
    }

    Booking {
        uuid id PK
        uuid asset_id FK
        uuid employee_id FK
        timestamp start_time
        timestamp end_time
        text purpose
        enum status "Upcoming | Active | Completed | Cancelled"
        timestamp created_at
        timestamp updated_at
    }

    AuditCycle {
        uuid id PK
        string name
        text description
        enum status "Planned | InProgress | Completed"
        date start_date
        date end_date
        uuid created_by FK
        timestamp created_at
        timestamp updated_at
    }

    AuditItem {
        uuid id PK
        uuid audit_cycle_id FK
        uuid asset_id FK
        uuid assigned_to FK
        enum status "Pending | Verified | Discrepancy"
        text notes
        timestamp verified_at
    }

    DiscrepancyReport {
        uuid id PK
        uuid audit_item_id FK
        uuid reported_by FK
        text description
        string expected_status
        string actual_status
        bool is_resolved
        uuid resolved_by FK
        text resolution_notes
        timestamp created_at
        timestamp updated_at
    }

    Notification {
        uuid id PK
        uuid employee_id FK
        string title
        text message
        enum type "Info | Warning | Action"
        bool is_read
        timestamp created_at
    }

    ActivityLog {
        uuid id PK
        uuid actor_id FK
        string action
        string entity_type
        uuid entity_id
        text details
        timestamp created_at
    }

    AssetDocument {
        uuid id PK
        uuid asset_id FK
        string file_name
        string file_url
        uuid uploaded_by FK
        timestamp created_at
        timestamp updated_at
    }
```

### Entity Descriptions

| Entity | Description |
|---|---|
| **Employee** | System users with one of four roles. Belongs to a department. |
| **Department** | Organizational unit grouping employees and assets. Can be deactivated. |
| **Category** | Classification label for assets (e.g., Laptops, Furniture, Vehicles). |
| **Asset** | A trackable item with lifecycle status, purchase/warranty info, and ownership. |
| **Allocation** | Represents an asset being assigned to an employee with a return deadline. |
| **TransferRequest** | An employee-initiated request to transfer an allocated asset to another employee. |
| **MaintenanceRequest** | A request to service/repair an asset, flowing through an approval pipeline. |
| **Booking** | A time-bound reservation of a shared asset/resource. |
| **AuditCycle** | A periodic physical verification campaign containing many audit items. |
| **AuditItem** | A single asset to be verified during an audit cycle. |
| **DiscrepancyReport** | Documents a mismatch found during audit verification. |
| **Notification** | An in-app alert sent to an employee about system events. |
| **ActivityLog** | An immutable record of actions performed in the system. |
| **AssetDocument** | A file (invoice, warranty certificate, etc.) attached to an asset. |

---

## Authentication & Authorization

### Auth Flow

```mermaid
sequenceDiagram
    participant U as User
    participant FE as React App
    participant API as Go Backend
    participant DB as PostgreSQL

    Note over U,DB: Signup / Login
    U->>FE: Enter credentials
    FE->>API: POST /api/auth/login
    API->>DB: Find employee by email
    DB-->>API: Employee record
    API->>API: Verify bcrypt hash
    API->>API: Generate JWT tokens
    API-->>FE: { accessToken, refreshToken }
    FE->>FE: Store in Zustand + localStorage

    Note over U,DB: Authenticated Request
    U->>FE: Navigate to page
    FE->>API: GET /api/v1/assets<br/>Authorization: Bearer <accessToken>
    API->>API: Validate JWT, extract claims
    API-->>FE: 200 OK + data

    Note over U,DB: Token Refresh (on 401)
    FE->>API: Any request → 401 Unauthorized
    FE->>API: POST /api/auth/refresh<br/>{ refreshToken }
    API->>API: Validate refresh token
    API-->>FE: { accessToken (new) }
    FE->>FE: Update Zustand store
    FE->>API: Retry original request
```

**JWT Token Details:**
- **Access Token** — 15-minute expiry. Contains claims: `employee_id`, `email`, `name`, `role`, `department_id`
- **Refresh Token** — 7-day expiry. Used to obtain new access tokens without re-login
- **Password Storage** — bcrypt hashing via `golang.org/x/crypto`

### Role-Based Access Control

```mermaid
graph TD
    subgraph Roles["👥 User Roles"]
        ADMIN["🔴 Admin"]
        AM["🟠 Asset Manager"]
        DH["🟡 Department Head"]
        EMP["🟢 Employee"]
    end

    subgraph Permissions["🔐 Permission Groups"]
        FULL["Full System Access"]
        ASSET_WRITE["Asset Registration<br/>& Document Upload"]
        MAINT_APPROVE["Maintenance<br/>Approval Workflow"]
        ALLOC_MANAGE["Allocation<br/>Create & Return"]
        TRANSFER_APPROVE["Transfer<br/>Approval"]
        DISCREP_RESOLVE["Discrepancy<br/>Resolution"]
        REPORTS["Reports &<br/>Activity Logs"]
        DEPT_MGMT["Department, Employee<br/>& Category CRUD"]
        AUDIT_MGMT["Audit Cycle<br/>Management"]
        READ_ALL["Read Assets, Categories<br/>Dashboard, Notifications"]
        BOOKING["Resource Booking<br/>& Maintenance Requests"]
        TRANSFER_REQ["Create Transfer<br/>Requests"]
    end

    ADMIN --> FULL
    FULL --> DEPT_MGMT
    FULL --> AUDIT_MGMT
    FULL --> ASSET_WRITE
    FULL --> MAINT_APPROVE
    FULL --> ALLOC_MANAGE
    FULL --> DISCREP_RESOLVE
    FULL --> REPORTS

    AM --> ASSET_WRITE
    AM --> MAINT_APPROVE
    AM --> ALLOC_MANAGE
    AM --> TRANSFER_APPROVE
    AM --> DISCREP_RESOLVE
    AM --> REPORTS

    DH --> ALLOC_MANAGE
    DH --> TRANSFER_APPROVE
    DH --> REPORTS

    EMP --> READ_ALL
    EMP --> BOOKING
    EMP --> TRANSFER_REQ

    style ADMIN fill:#dc2626,stroke:#fca5a5,color:#fff
    style AM fill:#ea580c,stroke:#fdba74,color:#fff
    style DH fill:#ca8a04,stroke:#fde047,color:#000
    style EMP fill:#16a34a,stroke:#86efac,color:#fff
```

**Detailed Permission Matrix:**

| Capability | Admin | AssetManager | DepartmentHead | Employee |
|---|:---:|:---:|:---:|:---:|
| Department CRUD | ✅ | ❌ | ❌ | ❌ |
| Category CRUD | ✅ | ❌ | ❌ | ❌ |
| Employee Management | ✅ | ❌ | ❌ | ❌ |
| Audit Cycle Create/Close | ✅ | ❌ | ❌ | ❌ |
| Assign Auditors | ✅ | ❌ | ❌ | ❌ |
| Register Assets | ✅ | ✅ | ❌ | ❌ |
| Upload Documents | ✅ | ✅ | ❌ | ❌ |
| Maintenance Approval | ✅ | ✅ | ❌ | ❌ |
| Resolve Discrepancies | ✅ | ✅ | ❌ | ❌ |
| Create/Return Allocations | ✅ | ✅ | ✅ | ❌ |
| Approve/Reject Transfers | ❌ | ✅ | ✅ | ❌ |
| View Reports | ✅ | ✅ | ✅ | ❌ |
| View Activity Logs | ✅ | ✅ | ✅ | ❌ |
| View Assets & Dashboard | ✅ | ✅ | ✅ | ✅ |
| Create Bookings | ✅ | ✅ | ✅ | ✅ |
| Request Maintenance | ✅ | ✅ | ✅ | ✅ |
| Request Transfers | ✅ | ✅ | ✅ | ✅ |
| Notifications | ✅ | ✅ | ✅ | ✅ |

---

## API Reference

All endpoints are prefixed with `/api`. Authenticated endpoints require `Authorization: Bearer <token>` header.

### Auth Endpoints

| Method | Path | Auth | Description |
|---|---|---|---|
| `POST` | `/auth/signup` | ❌ | Register a new employee |
| `POST` | `/auth/login` | ❌ | Authenticate & get tokens |
| `POST` | `/auth/refresh` | ❌ | Refresh access token |
| `POST` | `/auth/forgot-password` | ❌ | Password reset (stub) |
| `GET` | `/auth/me` | ✅ | Get current user info |

### Asset Endpoints

| Method | Path | Roles | Description |
|---|---|---|---|
| `GET` | `/v1/assets` | Any | List assets (search, filter, paginate) |
| `GET` | `/v1/assets/:id` | Any | Get asset details |
| `GET` | `/v1/assets/:id/history` | Any | Get asset allocation/maintenance/booking history |
| `POST` | `/v1/assets` | Admin, AssetManager | Register a new asset |
| `POST` | `/v1/assets/:id/documents` | Admin, AssetManager | Upload a document to an asset |
| `GET` | `/v1/categories` | Any | List asset categories |

### Allocation & Transfer Endpoints

| Method | Path | Roles | Description |
|---|---|---|---|
| `GET` | `/v1/allocations/my` | Any | List current user's allocations |
| `GET` | `/v1/allocations/overdue` | Any | List overdue allocations |
| `POST` | `/v1/allocations` | Admin, AssetManager, DeptHead | Create new allocation |
| `POST` | `/v1/allocations/:id/return` | Admin, AssetManager, DeptHead | Return an allocated asset |
| `GET` | `/v1/transfers/pending` | Any | List pending transfer requests |
| `POST` | `/v1/transfers` | Any | Create a transfer request |
| `PATCH` | `/v1/transfers/:id/approve` | AssetManager, DeptHead | Approve a transfer |
| `PATCH` | `/v1/transfers/:id/reject` | AssetManager, DeptHead | Reject a transfer |

### Maintenance Endpoints

| Method | Path | Roles | Description |
|---|---|---|---|
| `POST` | `/v1/maintenance` | Any | Create maintenance request |
| `GET` | `/v1/maintenance` | Any | List maintenance requests |
| `PATCH` | `/v1/maintenance/:id/approve` | Admin, AssetManager | Approve request |
| `PATCH` | `/v1/maintenance/:id/reject` | Admin, AssetManager | Reject request |
| `PATCH` | `/v1/maintenance/:id/assign-technician` | Admin, AssetManager | Assign technician |
| `PATCH` | `/v1/maintenance/:id/start` | Admin, AssetManager | Mark work started |
| `PATCH` | `/v1/maintenance/:id/resolve` | Admin, AssetManager | Mark as resolved |

### Booking Endpoints

| Method | Path | Roles | Description |
|---|---|---|---|
| `GET` | `/v1/resources/:assetId/bookings` | Any | List bookings for a resource |
| `POST` | `/v1/bookings` | Any | Create a booking |
| `GET` | `/v1/bookings/my` | Any | List current user's bookings |
| `PATCH` | `/v1/bookings/:id/cancel` | Any | Cancel a booking |
| `PATCH` | `/v1/bookings/:id/reschedule` | Any | Reschedule a booking |

### Audit Endpoints

| Method | Path | Roles | Description |
|---|---|---|---|
| `GET` | `/v1/audit-cycles` | Any | List all audit cycles |
| `GET` | `/v1/audit-cycles/:id` | Any | Get audit cycle details |
| `GET` | `/v1/audit-cycles/:id/items` | Any | List items in an audit cycle |
| `PATCH` | `/v1/audit-items/:id` | Any | Update audit item status |
| `GET` | `/v1/discrepancy-reports` | Any | List discrepancy reports |
| `PATCH` | `/v1/discrepancy-reports/:id/resolve` | Admin, AssetManager | Resolve a discrepancy |
| `POST` | `/v1/audit-cycles` | Admin | Create an audit cycle |
| `POST` | `/v1/audit-cycles/:id/auditors` | Admin | Assign auditors to cycle |
| `PATCH` | `/v1/audit-cycles/:id/close` | Admin | Close an audit cycle |

### Dashboard & Reports Endpoints

| Method | Path | Roles | Description |
|---|---|---|---|
| `GET` | `/v1/dashboard/kpis` | Any | Get KPI metrics |
| `GET` | `/v1/dashboard/overdue` | Any | Get overdue allocations |
| `GET` | `/v1/dashboard/upcoming` | Any | Get upcoming maintenance |
| `GET` | `/v1/dashboard/activity` | Any | Get recent activity feed |
| `GET` | `/v1/reports/utilization` | Admin, AssetManager, DeptHead | Asset utilization report |
| `GET` | `/v1/reports/maintenance-frequency` | Admin, AssetManager, DeptHead | Maintenance frequency report |
| `GET` | `/v1/reports/retirement-watchlist` | Admin, AssetManager, DeptHead | Warranty expiry watchlist |
| `GET` | `/v1/reports/allocation-summary` | Admin, AssetManager, DeptHead | Allocation statistics |
| `GET` | `/v1/reports/booking-heatmap` | Admin, AssetManager, DeptHead | Booking usage heatmap |
| `GET` | `/v1/reports/export` | Admin, AssetManager, DeptHead | CSV export of reports |

### Notifications & Activity Log Endpoints

| Method | Path | Roles | Description |
|---|---|---|---|
| `GET` | `/v1/notifications` | Any | List user notifications |
| `GET` | `/v1/notifications/unread-count` | Any | Get unread notification count |
| `PATCH` | `/v1/notifications/:id/read` | Any | Mark notification as read |
| `PATCH` | `/v1/notifications/read-all` | Any | Mark all notifications as read |
| `GET` | `/v1/activity-logs` | Admin, AssetManager, DeptHead | List activity logs |

### Admin Endpoints

| Method | Path | Description |
|---|---|---|
| `GET` | `/v1/departments` | List all departments |
| `POST` | `/v1/departments` | Create a department |
| `PATCH` | `/v1/departments/:id` | Update a department |
| `PATCH` | `/v1/departments/:id/deactivate` | Deactivate a department |
| `POST` | `/v1/categories` | Create a category |
| `PATCH` | `/v1/categories/:id` | Update a category |
| `GET` | `/v1/employees` | List all employees |
| `PATCH` | `/v1/employees/:id` | Update employee details |
| `PATCH` | `/v1/employees/:id/role` | Change employee role |

---

## Business Workflows

### Asset Lifecycle

```mermaid
stateDiagram-v2
    [*] --> Available: Asset Registered
    Available --> Allocated: Allocation Created
    Allocated --> Available: Allocation Returned
    Allocated --> TransferPending: Transfer Requested
    TransferPending --> Allocated: Transfer Approved<br/>(new employee)
    TransferPending --> Allocated: Transfer Rejected<br/>(original employee)
    Available --> UnderMaintenance: Maintenance Approved
    UnderMaintenance --> Available: Maintenance Resolved
    Available --> Retired: Admin Action
    Available --> Lost: Audit Discrepancy
    Allocated --> Overdue: Scheduler Detects<br/>Past Return Date
    Overdue --> Available: Allocation Returned
```

### Allocation & Transfer Workflow

```mermaid
flowchart TD
    A["Manager creates allocation"] --> B["Asset status → Allocated"]
    B --> C["Employee uses asset"]
    C --> D{Action?}
    
    D -->|Return| E["Manager processes return"]
    E --> F["Asset status → Available"]
    
    D -->|Transfer| G["Employee creates transfer request"]
    G --> H["Status: Pending"]
    H --> I{Manager Decision}
    I -->|Approve| J["Old allocation returned<br/>New allocation created<br/>Asset reassigned"]
    I -->|Reject| K["Transfer rejected<br/>Original allocation unchanged"]
    
    D -->|Overdue| L["Scheduler auto-detects"]
    L --> M["Status → Overdue<br/>Notification sent"]
    M --> E

    style A fill:#3b82f6,stroke:#93c5fd,color:#fff
    style F fill:#22c55e,stroke:#86efac,color:#fff
    style J fill:#22c55e,stroke:#86efac,color:#fff
    style K fill:#ef4444,stroke:#fca5a5,color:#fff
    style M fill:#f59e0b,stroke:#fde047,color:#000
```

### Maintenance Workflow

```mermaid
flowchart LR
    A["🔧 Employee submits<br/>maintenance request"] --> B["📋 Status: Pending"]
    B --> C{Admin / Asset Manager}
    C -->|Approve| D["✅ Status: Approved<br/>Asset → UnderMaintenance"]
    C -->|Reject| E["❌ Status: Rejected"]
    D --> F["👷 Assign Technician"]
    F --> G["🔨 Status: InProgress"]
    G --> H["✅ Status: Resolved<br/>Asset → Available"]

    style A fill:#3b82f6,stroke:#93c5fd,color:#fff
    style B fill:#f59e0b,stroke:#fde047,color:#000
    style D fill:#8b5cf6,stroke:#c4b5fd,color:#fff
    style E fill:#ef4444,stroke:#fca5a5,color:#fff
    style F fill:#06b6d4,stroke:#67e8f9,color:#000
    style G fill:#f97316,stroke:#fdba74,color:#fff
    style H fill:#22c55e,stroke:#86efac,color:#fff
```

### Audit Workflow

```mermaid
flowchart TD
    A["Admin creates<br/>Audit Cycle"] --> B["Status: Planned"]
    B --> C["Admin assigns auditors"]
    C --> D["Auto-generate audit items<br/>(one per asset)"]
    D --> E["Status: InProgress"]
    E --> F["Auditors verify items"]
    
    F --> G{Physical Check}
    G -->|Match| H["Item: Verified ✅"]
    G -->|Mismatch| I["Item: Discrepancy ⚠️"]
    I --> J["Create Discrepancy Report"]
    J --> K["Admin/AssetManager<br/>resolves discrepancy"]
    
    H --> L{All items checked?}
    K --> L
    L -->|No| F
    L -->|Yes| M["Admin closes cycle"]
    M --> N["Status: Completed"]

    style A fill:#3b82f6,stroke:#93c5fd,color:#fff
    style H fill:#22c55e,stroke:#86efac,color:#fff
    style I fill:#f59e0b,stroke:#fde047,color:#000
    style J fill:#ef4444,stroke:#fca5a5,color:#fff
    style N fill:#22c55e,stroke:#86efac,color:#fff
```

### Booking Workflow

```mermaid
flowchart LR
    A["Employee selects<br/>resource & time slot"] --> B{Conflict?}
    B -->|Yes| C["❌ Booking rejected<br/>(time conflict)"]
    B -->|No| D["✅ Booking created<br/>Status: Upcoming"]
    D --> E["⏰ Scheduler:<br/>start_time reached"]
    E --> F["Status: Active"]
    F --> G["⏰ Scheduler:<br/>end_time reached"]
    G --> H["Status: Completed"]
    
    D --> I["Employee cancels"]
    I --> J["Status: Cancelled"]
    
    D --> K["Employee reschedules"]
    K --> B

    style C fill:#ef4444,stroke:#fca5a5,color:#fff
    style D fill:#3b82f6,stroke:#93c5fd,color:#fff
    style F fill:#f59e0b,stroke:#fde047,color:#000
    style H fill:#22c55e,stroke:#86efac,color:#fff
    style J fill:#6b7280,stroke:#d1d5db,color:#fff
```

---

## Background Jobs

The scheduler runs two periodic background tasks on startup:

```mermaid
flowchart TB
    subgraph Scheduler["⏱️ Background Scheduler"]
        direction TB
        
        subgraph Job1["Overdue Allocation Checker"]
            T1["Every 1 hour"] --> Q1["Query: Active allocations<br/>past expected_return"]
            Q1 --> U1["Update status → Overdue"]
            U1 --> N1["Send notification<br/>to employee"]
        end
        
        subgraph Job2["Booking Status Updater"]
            T2["Every 15 minutes"] --> Q2A["Query: Upcoming bookings<br/>past start_time"]
            Q2A --> U2A["Update status → Active"]
            T2 --> Q2B["Query: Active bookings<br/>past end_time"]
            Q2B --> U2B["Update status → Completed"]
        end
    end

    style Scheduler fill:#1e1b4b,stroke:#818cf8,color:#e2e8f0
    style Job1 fill:#1e3a5f,stroke:#60a5fa,color:#e2e8f0
    style Job2 fill:#1e3a5f,stroke:#60a5fa,color:#e2e8f0
```

---

## Frontend Pages

```mermaid
graph TD
    subgraph Public["🌐 Public Routes"]
        LOGIN["/login"]
        SIGNUP["/signup"]
    end

    subgraph App["🔒 Authenticated Routes (/_app)"]
        DASH["/ — Dashboard<br/>KPIs, Overdue, Upcoming, Activity"]
        ASSETS["/assets — Asset List<br/>Search, Filter, Paginate"]
        ASSET_DETAIL["/assets/:id — Asset Detail<br/>Tabs: Overview, History, Docs"]
        ASSET_REG["/assets/register — Register Asset"]
        ALLOC["/allocations — Allocations"]
        ALLOC_NEW["/allocations/new — New Allocation"]
        TRANSFERS["/transfers — Transfer Requests"]
        MAINT["/maintenance — Maintenance"]
        BOOKINGS["/bookings — Resource Bookings"]
        AUDIT["/audit — Audit Cycles"]
        AUDIT_DETAIL["/audit/:id — Audit Detail"]
        REPORTS["/reports — Analytics & Charts"]
        DEPTS["/departments — Departments"]
        EMPS["/employees — Employees"]
        NOTIF["/notifications — Notifications"]
        SETTINGS["/settings — User Settings"]
    end

    LOGIN -->|Auth Success| DASH
    SIGNUP -->|Auth Success| DASH
    ASSETS --> ASSET_DETAIL
    ASSETS --> ASSET_REG
    ALLOC --> ALLOC_NEW
    AUDIT --> AUDIT_DETAIL

    style Public fill:#1e293b,stroke:#f472b6,color:#e2e8f0
    style App fill:#0f172a,stroke:#6366f1,color:#e2e8f0
```

---

## Getting Started

### Prerequisites

- **Go** ≥ 1.25
- **Node.js** ≥ 22
- **pnpm** ≥ 10
- **PostgreSQL** ≥ 15 (or Docker)
- **Task** (task runner) — [Install instructions](https://taskfile.dev/docs/installation)

```bash
# Install Task (Debian/Ubuntu)
curl -1sLf 'https://dl.cloudsmith.io/public/task/task/setup.deb.sh' | sudo -E bash
apt install task

# Install Task (macOS)
brew install go-task/tap/go-task

# Install Task (Windows)
winget install Task.Task
```

### Local Development

```bash
# 1. Clone the repository
git clone https://github.com/saisrikardumpeti/odoo-hackathon-2026.git
cd odoo-hackathon-2026

# 2. Set up environment variables
cp .env.example .env
# Edit .env with your local PostgreSQL connection string

# 3. Install dependencies
task setup:backend    # Runs `go mod tidy`
task setup:frontend   # Runs `pnpm install`

# 4. (Optional) Seed the database
task seed:db          # Runs seed.go to populate sample data

# 5. Start development servers
task run:backend:dev   # Starts Go server with Air hot-reload on :8000
task run:frontend:dev  # Starts Vite dev server on :5173
```

### Docker Development

```bash
# Start all services (PostgreSQL + Backend + Frontend)
docker compose up --build

# Services will be available at:
#   Frontend:  http://localhost:5173
#   Backend:   http://localhost:8000
#   Database:  localhost:5432
```

```mermaid
graph LR
    subgraph DockerCompose["🐳 docker compose up"]
        DB["postgres:15-alpine<br/>:5432"]
        BE["go_backend<br/>:8000<br/>(Air hot-reload)"]
        FE["vite_frontend<br/>:5173<br/>(pnpm dev)"]
    end

    FE -->|depends_on| BE -->|depends_on| DB
    BE -.->|"volume mount"| SRC1["./apps/backend"]
    FE -.->|"volume mount"| SRC2["./apps/frontend"]
    DB -.->|"named volume"| VOL["pgdata"]

    style DockerCompose fill:#0f172a,stroke:#6366f1,color:#e2e8f0
```

---

## Environment Variables

| Variable | Description | Example |
|---|---|---|
| `DATABASE_URL` | PostgreSQL connection string | `postgresql://odoo:odoo@db:5432/odoo-hackathon` |
| `BACKEND_URL` | Backend API URL (used by frontend container) | `http://backend:8080` |

---

## Task Runner Commands

| Command | Description |
|---|---|
| `task setup:backend` | Install Go backend dependencies (`go mod tidy`) |
| `task setup:frontend` | Install frontend dependencies (`pnpm install`) |
| `task seed:db` | Seed the database with sample data |
| `task run:backend:dev` | Start backend with Air hot-reload |
| `task run:frontend:dev` | Start frontend Vite dev server |

---

<p align="center">
  Built with ❤️ for the <strong>Odoo Hackathon 2026</strong>
</p>
