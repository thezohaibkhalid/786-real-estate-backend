# 786 Real Estate backend

Go Fiber v2 + PostgreSQL foundation for the existing public Next.js website and admin panel. Fiber v2.52.15 and pgx v5.11.0 are pinned to the locally available versions; this starter does not claim to use Fiber v3. Go 1.26+ required. The module path is a placeholder: replace `example.com/786-real-estate/backend` and its imports when the repository URL is decided.

## Local start

```sh
cp .env.example .env
# Edit .env if your PostgreSQL credentials or admin port differ.
docker compose up -d db
go mod download
make migrate
# For live reload during development:
make dev    # or: air
# Or to run once:
make run
```

The development environment is loaded from `.env`; explicitly exported environment variables take precedence. Compose is for local development only; use separately managed credentials and TLS in production. No database is provisioned or migrated automatically when the API starts.
The local Postgres container is published on `127.0.0.1:5433` to avoid clashing with other local Postgres projects that may already use `5432`.

- `GET /health/live`: process health.
- `GET /health/ready`: database connectivity; returns 503 on failure. It does not verify schema version.
- `GET /api`: service metadata with the project's data/meta envelope.
- Public read endpoints from `API_REQUIREMENTS.md`: properties, projects, cities, areas, blog, SEO pages, FAQs, testimonials, legal documents, offices, settings and sitemap.
- Public write endpoints: `POST /api/leads` and `POST /api/property-alerts`, with validation, JSON error fields and an in-process 10/hour/IP limiter.
- Admin auth endpoints: `POST /api/admin/auth/login`, `POST /api/admin/auth/logout`, `GET /api/admin/auth/me`, `POST /api/admin/auth/forgot-password`, and `POST /api/admin/auth/reset-password`.
- Admin dashboard and leads endpoints: `GET /api/admin/dashboard`, `GET /api/admin/leads`, `GET /api/admin/leads/:id`, `PATCH /api/admin/leads/:id`, and `GET /api/admin/cities`.
- Development admin login: `admin@786realestate.pk` / `ChangeMe786!`. Change this before production.
- Other paths: JSON 404. Full admin CRUD, uploads, and all role-specific write endpoints are still being implemented.

## Layout

- `cmd/api`: startup, signals and graceful shutdown.
- `cmd/migrate`: explicit, transactional forward migrations with locking and checksum validation.
- `internal/config`: validated environment configuration.
- `internal/database`: pgx pool with bounded connection times.
- `internal/httpapi`: routes, error envelopes, request IDs, recovery, logging and CORS.
- `migrations`: embedded SQL, applied in filename order.

Add domain packages under `internal` as endpoints are implemented. Handlers should validate transport input, services implement business rules, and repositories execute parameterized SQL. Use transactions for nested collection replacement and lead/alert creation. Never return raw database rows containing internal fields.

## Database baseline

`0001_initial.sql` is copied from the project's existing schema, with only its top-level BEGIN/COMMIT removed so the runner owns the transaction (including views). It is intended for an EMPTY database. Do not apply it to an existing manually-created schema. Existing databases need an explicit baseline reconciliation first. Applied migrations must never be edited; add numbered SQL files instead. Rollbacks require a reviewed forward correction or restoring a backup; there is no destructive down command.

Known original schema/API inconsistencies remain deliberate follow-up work: city/area consistency, persisted area slug/publication support, SEO tags, scheduled publishing, project archival and currency code route keys. The public area endpoint currently derives the slug from `areas.name` and only serves areas that have at least one indexed SEO page. Baseline public content and the development owner account are seeded by migration.

## Validation

```sh
make check
make test
make build
```

HTTP tests run without PostgreSQL using a health dependency fake. A real PostgreSQL migration and readiness smoke test should also be run before deployment. Docker builds are optional.

Public CORS permits anonymous requests. Admin CORS permits only ADMIN_ORIGIN plus the equivalent localhost/127.0.0.1 development origin, with credentials; CORS is not authorization. Forgot-password emails use `EMAIL_SERVER_*` SMTP settings and store only hashed reset tokens. Before production, replace the seeded password, set a strong `SESSION_SECRET`, use production SMTP credentials, add login throttling and CSRF/origin protection, and replace the in-process public write limiter with a distributed limiter before running multiple API instances. Logs omit bodies, URLs/query strings and database credentials.

For a container, set HTTP_ADDR=0.0.0.0:8080 and supply DATABASE_URL/ADMIN_ORIGIN at runtime. Run `/app/migrate` as a separate deployment step before starting `/app/api`.
