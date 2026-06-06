# Biolynq Go Backend

Biolynq is a high-performance, Linktree-inspired link aggregation and creator dashboard platform. This repository contains the Go-based backend microservices, featuring a REST API server, background task queue processor, normalized analytics database structure, and multi-provider authentication.

---

## Technical Stack

* **Language**: Go (v1.26+)
* **HTTP Framework**: [Fiber v3](https://gofiber.io/)
* **ORM**: [GORM](https://gorm.io/) (PostgreSQL dialect)
* **Database**: PostgreSQL (v17)
* **Caching & Queue**: Redis (v7)
* **Background Task Queue**: [Asynq](https://github.com/hibiken/asynq) (Redis-backed)
* **Migration Management**: [Atlas CLI](https://atlasgo.io/)

---

## Features

1. **Authentication Flows**
   * Passwordless Email OTP authentication (sign-up, login, verify).
   * Google OAuth 2.0 integration.
   * State-aware session management with JWT & sliding tokens.
2. **Profile & Customization**
   * Unique username check and reservation.
   * Profile customizations (Display Name, Bio, Themes).
   * Profile routes automatically return active links sorted by order/position.
3. **Links Management**
   * Consolidated standard & social links under a single `links` table using an `is_social` flag.
   * Support for CRUD operations on links with custom link reordering.
4. **Asynchronous Analytics Module**
   * **Redirection & Clicks**: Public clicks are routed through a redirection controller, issuing quick `303 See Other` responses while enqueuing events in Redis queues.
   * **Profile Views**: Page views are tracked asynchronously when a profile is publicly retrieved.
   * **ACID Transactions**: A normalized GORM schema separating `analytic_events` and `visitor_metadata` (IP, UserAgent, Browser, OS, Device, City, Country) using single database transaction locking to optimize insert speeds and prevent race conditions.
5. **Dashboard Analytics APIs**
   * **Overview**: Aggregates total views, link clicks, dynamic click-through-rates (CTR), unique visitor counts, and daily time-series chart statistics (views vs clicks over the last 30 days, zero-padded for days without traffic).
   * **Links Stats**: Grouped click counts and CTR rates per standard/social link.
   * **Demographics Breakdown**: Aggregated statistics of visitors grouped by Device, Browser, OS, Referrer, Country, and City.

---

## Project Structure

```text
├── cmd/
│   ├── server/             # API Server entrypoint
│   └── worker/             # Background task worker entrypoint
├── docker/                 # Container definition files
├── internal/
│   ├── analytics/          # Analytics endpoints, aggregates & redirection
│   ├── auth/               # OTP & OAuth controllers & services
│   ├── bootstrap/          # Fiber initialization & router wiring
│   ├── config/             # Environment configuration mapping
│   ├── database/           # PostgreSQL & Redis connection pools
│   ├── links/              # Link CRUD controllers & repositories
│   ├── middleware/         # Authentication and request tracking middleware
│   ├── models/             # Database/GORM schemas (User, Link, AnalyticEvent, etc.)
│   ├── templates/          # HTML email templates
│   ├── users/              # Profiles, Users, Sessions & User Handlers
│   ├── utils/              # Email client wrappers & validation helpers
│   └── worker/             # Task enqueuers & processors (Asynq)
├── migrations/             # Atlas SQL migration files
├── docker-compose.yml      # Service dependencies setup
├── Makefile                # Useful CLI shortcuts
└── go.mod                  # Dependency file
```

---

## Setup & Execution

### Prerequisites
* Go v1.26 or higher
* Docker and Docker Compose
* Atlas CLI (optional, for custom migration generation)

### 1. Environment Configuration
Create a `.env` file in the root directory:
```env
PORT=8000
DB_URL=postgres://postgres:postgres@localhost:5432/biolynq_db?sslmode=disable
REDIS_URL=localhost:6379
JWT_SECRET=your-secret-key
GOOGLE_REDIRECT_URL=http://localhost:8000/api/v1/auth/google/callback
FRONTEND_URL=http://localhost:3000
RESEND_KEY=re_your_api_key_here
```

### 2. Start Services via Docker Compose
To build and spin up the PostgreSQL database, Redis instance, API server, and background worker:
```bash
make docker-up
```
This launches the containers:
* **Server**: listens on `http://localhost:8000`
* **Worker**: background job processing engine
* **Database (PostgreSQL)**: port `5432`
* **Redis**: port `6379`

To stop services:
```bash
make docker-down
```

### 3. Running Locally (Development Mode)
If you prefer running services directly on your machine:
```bash
# Run API Server
make run-server

# Run Worker Processor
make run-worker

# Run tests
make test
```

---

## Database Migrations (Atlas)

Database schemas are defined inside the `internal/models` Go code. Migration generation uses Atlas to diff GORM models against the local Postgres environment:

```bash
# Generate a new migration file after modifying models
make migration-gen name=add_your_feature_fields

# Apply migrations manually
make migration-apply

# Verify migration status
make migration-status
```

---

## API Endpoints List

### Public Routes
* `GET /api/v1/public/:username` - Retrieve public profile layout (includes profile details and active links).
* `GET /api/v1/public/links/:linkID/redirect` - Public link click tracking and 303 redirection destination.

### Authentication Routes
* `POST /api/v1/auth/otp/send` - Send verification OTP to email.
* `POST /api/v1/auth/otp/verify` - Verify OTP and receive authentication session token.
* `GET /api/v1/auth/google/login` - Initiate Google OAuth login.
* `GET /api/v1/auth/google/callback` - OAuth authorization redirect target.
* `POST /api/v1/auth/register/complete` - Save username & complete registration.
* `POST /api/v1/auth/check-username` - Check availability of username.
* `POST /api/v1/auth/logout` - Invalidate active user session (authenticated).

### Authenticated Links Routes
* `GET /api/v1/links` - Get all links owned by authenticated user.
* `POST /api/v1/links` - Create a new link (standard/social).
* `GET /api/v1/links/:linkID` - Get details of specific link.
* `PUT /api/v1/links/:linkID` - Update link configuration (title, URL, position, visibility, etc.).
* `DELETE /api/v1/links/:linkID` - Delete a link.

### Authenticated Creator Analytics Routes
* `GET /api/v1/analytics/overview` - Combined views, clicks, CTR, unique visitor counts, and 30-day stats daily time-series maps.
* `GET /api/v1/analytics/links` - Link-by-link total click aggregates and specific click-through-rates.
* `GET /api/v1/analytics/breakdown` - Visitor details classified by Devices, OS, Browsers, Referrers, Cities, and Countries.
