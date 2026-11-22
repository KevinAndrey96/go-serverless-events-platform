# Serverless Events Platform

[![CI](https://github.com/KevinAndrey96/serverless-events-platform/actions/workflows/ci.yml/badge.svg)](https://github.com/KevinAndrey96/serverless-events-platform/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/Go-1.20+-00ADD8?style=flat&logo=go)](https://golang.org)
[![AWS Lambda](https://img.shields.io/badge/AWS-Lambda-FF9900?style=flat&logo=awslambda)](https://aws.amazon.com/lambda/)
[![DynamoDB](https://img.shields.io/badge/Amazon-DynamoDB-4053D6?style=flat&logo=amazondynamodb)](https://aws.amazon.com/dynamodb/)
[![Serverless](https://img.shields.io/badge/Serverless-Framework-FD5750?style=flat&logo=serverless)](https://www.serverless.com/)
[![Laravel](https://img.shields.io/badge/Laravel-9.x-FF2D20?style=flat&logo=laravel)](https://laravel.com)
[![Vite](https://img.shields.io/badge/Vite-3.x-646CFF?style=flat&logo=vite)](https://vitejs.dev/)
[![QR Scanner](https://img.shields.io/badge/QR%20Scanner-Instascan.js-00A98F?style=flat)](https://github.com/schmich/instascan)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

An enterprise-grade, event-driven full-stack platform engineered as a polyglot monorepo. The architecture combines a high-performance **Go (Golang) AWS Serverless microservices backend** with a **Laravel 9 & Vite full-stack administrative and public portal**.

The platform powers the end-to-end lifecycle of events and ticketing: organizer governance, public attendee self-registration, viral friend referrals, automated vector QR code pass generation and PDF delivery, cron-scheduled 24-hour reminder dispatches via TLS SMTP, and physical venue check-in via an in-browser live camera QR scanner.

---

## System Architecture

```mermaid
flowchart TD
    subgraph Clients["Clients & Entry Devices"]
        Admin["Organizer / Admin"]
        PublicUser["Public Attendee"]
        DoorStaff["Door Staff / Scanner<br/>(Camera Device)"]
    end

    subgraph FrontendApp["Frontend & Web Portal (Laravel 9 + Vite)"]
        WebRoutes["Web Router & Middleware"]
        InvokableControllers["Single-Action Controllers"]
        UseCases["Hexagonal Domain Use Cases"]
        EventRepo["EventRepository (HTTP Client)"]
        DocService["DomPDF & SimpleQrCode Engine"]
    end

    subgraph ServerlessBackend["Backend Microservices (Go / AWS Lambda / provided.al2)"]
        APIGW["Amazon API Gateway"]
        CreateLambda["create-event-api<br/>(POST /v1)"]
        ShowLambda["show-events-api<br/>(GET /v1?param=...)"]
        StatusLambda["change-status-api<br/>(POST / PATCH /v1)"]
        PayLambda["change-pay-status-api<br/>(POST / PATCH /v1)"]
        UsersLambda["event-users-api<br/>(GET /v1?eventid=...)"]
        ReminderLambda["reminder-events-api<br/>(Scheduled Cron Worker)"]
    end

    subgraph CloudInfra["Cloud Data & Event Infrastructure"]
        DDB[("Amazon DynamoDB<br/>Single-Table Design<br/>(sk-index, eventid-index)")]
        EventBridge["Amazon EventBridge<br/>(6-Hour Scheduled Rule)"]
        MailServer["Outbound Mail Server<br/>(TLS SMTP)"]
    end

    Admin -->|Manage Events & Attendees| WebRoutes
    PublicUser -->|Register & Upload Receipt| WebRoutes
    DoorStaff -->|Live Camera QR Scan| WebRoutes

    WebRoutes --> InvokableControllers
    InvokableControllers --> UseCases
    UseCases --> EventRepo
    InvokableControllers --> DocService

    DocService -->|Attach PDF Pass| MailServer

    EventRepo -->|REST / JSON| APIGW
    APIGW --> CreateLambda
    APIGW --> ShowLambda
    APIGW --> StatusLambda
    APIGW --> PayLambda
    APIGW --> UsersLambda

    CreateLambda -->|PutItem| DDB
    ShowLambda -->|Query sk-index| DDB
    StatusLambda -->|UpdateItem #S = :s| DDB
    PayLambda -->|UpdateItem #PS = :ps| DDB
    UsersLambda -->|Query eventid-index| DDB

    EventBridge -->|Trigger Cron| ReminderLambda
    ReminderLambda -->|Query Users & Events| DDB
    ReminderLambda -->|Send 24h Reminder| MailServer
```

---

## End-to-End User & Ticketing Flow

```mermaid
sequenceDiagram
    autonumber
    actor Organizer as Event Organizer
    actor Attendee as Public Attendee
    actor Staff as Door Staff (Scanner)
    participant Portal as Laravel Web Portal
    participant API as AWS Lambda (Go)
    participant DDB as Amazon DynamoDB
    participant Cron as EventBridge / Reminder
    participant Mail as SMTP Mail Server

    Organizer->>Portal: Create new Event (capacity, price, date, time)
    Portal->>API: POST /v1 (Event metadata)
    API->>DDB: PutItem (PK: Event#UUID, SK: METADATA#EVENT)
    DDB-->>API: 200 OK
    API-->>Portal: Event Created

    Attendee->>Portal: Register for event & submit details
    Portal->>API: POST /v1 (User attendee metadata)
    API->>DDB: PutItem (PK: User#UUID, SK: METADATA#USER, eventid: Event#UUID)
    Portal->>Portal: Generate vector QR & render PDF ticket (DomPDF)
    Portal->>Mail: Send access ticket email with PDF attachment
    Mail-->>Attendee: Receive Ticket with QR Code

    Note over Cron,Reminder: Automated Reminder Cycle (Every 6h)
    Cron->>API: Invoke reminder-events-api
    API->>DDB: Query sk-index (METADATA#USER) + Fetch Event Details
    API->>API: Filter events starting in <= 24h (COT -5)
    API->>Mail: Dispatch 24h reminder emails via TLS SMTP
    Mail-->>Attendee: Receive Event Reminder

    Note over Staff,DDB: Physical Venue Check-in
    Attendee->>Staff: Present QR Code (mobile or printed)
    Staff->>Portal: Open /scanner-show (live camera stream)
    Portal->>Portal: Instascan.js decodes QR code token
    Portal->>API: POST /v1 (Update status: 'attend')
    API->>DDB: UpdateItem (SET #S = :s)
    API-->>Portal: Check-in Confirmed
    Portal-->>Staff: Visual check-in success banner
```

---

## Monorepo Project Structure

```
serverless-events-platform/
├── backend/                       # Go AWS Serverless Microservices Platform
│   ├── pkg/
│   │   ├── config/                # AWS SDK session singleton & dynamic env configuration
│   │   ├── models/                # Strongly-typed domain representations & DDB mappings
│   │   └── response/              # Standard CORS & structured JSON API responses
│   ├── create/                    # POST /v1: Event & ticket registration
│   │   ├── serverless.yml
│   │   └── v1/main.go
│   ├── show/                      # GET /v1: Catalog search on GSI (sk-index)
│   │   ├── serverless.yml
│   │   └── v1/main.go
│   ├── change_status/             # POST / PATCH /v1: Event lifecycle transition
│   │   ├── serverless.yml
│   │   └── v1/main.go
│   ├── change_pay_status/         # POST / PATCH /v1: Payment reconciliation
│   │   ├── serverless.yml
│   │   └── v1/main.go
│   ├── event-users/               # GET /v1: Attendee query on GSI (eventid-index)
│   │   ├── serverless.yml
│   │   └── v1/main.go
│   ├── reminder/                  # Scheduled Worker: 24h TLS SMTP reminder cron
│   │   ├── serverless.yml
│   │   └── v1/main.go
│   ├── Makefile                   # Service-level Go compilation & test rules
│   ├── go.mod
│   └── go.sum
├── frontend/                      # Laravel 9 + Vite Full-Stack Application
│   ├── app/
│   │   ├── Http/Controllers/      # Single-Action Invokable Controllers
│   │   │   ├── Events/            # Event creation, catalog & attendees
│   │   │   ├── Records/           # Public enrollment & friend referrals
│   │   │   ├── Scanner/           # Live video camera QR validator
│   │   │   └── Attendee/          # Payment audit & approval
│   │   ├── Repositories/          # Domain contracts & HTTP client implementations
│   │   └── UsesCases/             # Hexagonal application business logic
│   ├── config/services.php        # Centralized microservices endpoint map
│   ├── resources/views/           # Blade templates (Dashboard, Scanner, PDF Pass)
│   ├── routes/web.php             # Route definitions
│   └── package.json
├── Makefile                       # Root monorepo build & test orchestrator
└── README.md
```

---

## DynamoDB Single-Table Schema & Access Patterns

The backend consolidates all domain entities into a single DynamoDB table to achieve predictable sub-10ms performance at scale.

### Key Structure & Secondary Indexes

| Entity | Partition Key (`pk`) | Sort Key (`sk`) | GSI: `sk-index` (PK) | GSI: `eventid-index` (PK) | Attributes |
|---|---|---|---|---|---|
| **Event Metadata** | `Event#<UUID>` | `METADATA#EVENT` | `METADATA#EVENT` | — | `name`, `capacity`, `price`, `date`, `hour`, `status`, `bank`, `account` |
| **Attendee / Ticket** | `User#<UUID>` | `METADATA#USER` | `METADATA#USER` | `Event#<UUID>` | `name`, `lastname`, `email`, `phone`, `status`, `paymentstatus`, `scholarship` |

### Access Pattern Matrix

| Access Pattern | Microservice | Query Target | Method |
|---|---|---|---|
| Register Event or Attendee | `create-event-api` | Table | `PutItem` |
| List Events Catalog | `show-events-api` | `sk-index` (`sk = METADATA#EVENT`) | `Query` |
| Retrieve Attendees for an Event | `event-users-api` | `eventid-index` (`eventid = Event#<UUID>`) | `Query` |
| Transition Event State | `change-status-api` | Table (`pk`, `sk`) | `UpdateItem` |
| Confirm Payment Status | `change-pay-status-api` | Table (`pk`, `sk`) | `UpdateItem` |
| Evaluate Upcoming Reminders | `reminder-events-api` | `sk-index` (`sk = METADATA#USER`) + `GetItem` Event | `Query` + `GetItem` |

---

## Microservices Integration Contracts

The frontend interacts with the Go backend via HTTP/REST contracts defined in `frontend/config/services.php`:

| Operation | Environment Variable | HTTP Method | Payload / Query |
|---|---|---|---|
| Catalog Query | `URL_SHOW_EVENTS_USERS` | `GET` | `?param=EVENT` |
| Create Event / Attendee | `URL_CREATE_EVENTS` | `POST` | Entity JSON payload |
| Update Status | `URL_CHANGE_STATUS_USERS_EVENTS` | `POST` / `PATCH` | `{"pk": "...", "sk": "...", "status": "..."}` |
| Query Event Attendees | `URL_EVENT_USERS` | `GET` | `?param={eventID}` |
| Reconcile Payment | `URL_CHANGE_PAY_STATUS` | `POST` / `PATCH` | `{"pk": "...", "sk": "...", "status": "PAID"}` |

---

## Engineering Highlights & Design Patterns

1. **Polyglot Monorepo Architecture:**
   - Single source of truth with unified root orchestration (`Makefile`).
   - Clean domain boundaries: Go for high-throughput, low-latency microservices and scheduled background workers; Laravel for rapid, expressive web UI, validation, and PDF rendering.

2. **Decoupled Hexagonal / Clean Architecture:**
   - The Laravel web portal defines `EventRepositoryInterface`, ensuring controllers and use cases depend exclusively on abstractions.
   - Resilient HTTP communication: includes timeouts (10s), error handling, and structured logging (`Log::error`).

3. **12-Factor App & Zero Hardcoded Secrets:**
   - Fully sanitized codebase: all database credentials, API endpoints, and mail credentials are injected via environment variables (`.env`).
   - Automated dry-run mode: the reminder background worker safely skips outbound dispatches when `SMTP_PASS` is omitted in staging environments without panicking.

4. **Hardware Integration in Modern Web:**
   - Client-side live camera stream decoding powered by `Instascan.js`.
   - Native device camera detection with mobile rear-camera selection for contactless gate check-in.

---

## Quickstart & Local Setup

### 1. Root Monorepo Commands
```bash
# Start local DynamoDB, Admin UI, MySQL & Mailpit
make docker-up

# Run all unit test suites
make test

# Compile all Go Lambda binaries
make build

# Stop local Docker environment
make docker-down
```

### 2. Local Infrastructure Services (Docker Compose)
| Service | Container Port | Host Port | Purpose / Dashboard |
| :--- | :---: | :---: | :--- |
| **DynamoDB Local** | `8000` | `8000` | Local AWS DynamoDB engine for offline Lambda invocation |
| **DynamoDB Admin** | `8001` | `8001` | Web UI at [http://localhost:8001](http://localhost:8001) to inspect single-table items |
| **MySQL 8.0** | `3306` | `3306` | Relational store for the Laravel administrative portal |
| **Mailpit** | `1025 / 8025` | `1025 / 8025` | SMTP sink + Webmail at [http://localhost:8025](http://localhost:8025) for reminder emails |

### 3. Backend Deployment (Go / AWS Lambda)
Prerequisites: Go 1.20+, Serverless Framework v3 (`npm install -g serverless`), configured AWS CLI.
```bash
cd backend/create
serverless deploy --stage dev
```
*(Repeat for each microservice under `backend/`)*

### 3. Frontend Setup (Laravel 9 / Vite)
Prerequisites: PHP 8.1+, Composer, Node.js 18+.
```bash
cd frontend

# Install dependencies
composer install
npm install

# Environment configuration
cp .env.example .env
php artisan key:generate

# Run migrations & seeders
php artisan migrate --seed

# Build assets & start dev server
npm run build
php artisan serve
```
Open `http://localhost:8000` to access the portal.

---

## Author
**Kevin Andrey Herrera Silva**
- GitHub: [@KevinAndrey96](https://github.com/KevinAndrey96)
- LinkedIn: [Kevin Andrey Herrera Silva](https://www.linkedin.com/in/kevinandrey96/)
