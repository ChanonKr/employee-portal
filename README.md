# Employee Profile Portal

Employees view and update their own profile: photo, contact details (one field at a
time), and skills, with a change history for the last N days. Admins manage the skill
catalog and the history window.

**Backend:** Go 1.27 · Gin · PostgreSQL 17 · pgx + sqlc · golang-migrate
**Frontend:** React + Vite (in progress)

Full design and trade-offs: [docs/DESIGN.md](docs/DESIGN.md)

## Run it

Requires Docker.

```bash
make up          # creates .env with a random JWT secret, then starts postgres → migrations → API
curl localhost:8080/api/v1/healthz
```

| Account     | Password   | What it shows                                              |
|-------------|------------|------------------------------------------------------------|
| `EMP-00001` | `demo1234` | Complete profile; history inside and outside the 7-day window |
| `EMP-00002` | `demo1234` | Empty profile (first-time experience)                      |
| `ADMIN-001` | `demo1234` | Admin role                                                 |

Demo data is fictional and lives in migration `000003_seed_demo` (local development only).

Other commands: `make help`.

## API (`/api/v1`)

| Method | Path | Notes |
|---|---|---|
| POST | `/auth/login` | `{employeeCode, password}` → HttpOnly session cookie |
| POST | `/auth/logout` | |
| GET  | `/auth/me` | |
| POST | `/auth/change-password` | `{currentPassword, newPassword}`; required after a temporary password |
| GET  | `/profile` | `ETag: "<version>"` |
| PUT  | `/profile/contact/:field` | `mobile` \| `email` \| `address`, body `{value}`, header `If-Match` |
| PUT  | `/profile/photo` | multipart `file` (JPEG/PNG/WebP ≤ 5 MB), `If-Match` |
| PUT  | `/profile/skills` | `{skillIds: number[]}`, `If-Match` |
| GET  | `/profile/history` | `?field=&page=&pageSize=` |
| GET  | `/skills` | active skill catalog |
| GET  | `/files/photos/:name` | |
| GET  | `/healthz` | |

Admin only (`/api/v1/admin`, checked in the router and again in each use case):

| Method | Path | Notes |
|---|---|---|
| GET | `/stats` | employees, complete profiles, top skills, recent changes |
| GET | `/employees` | `?q=&page=&pageSize=` |
| GET | `/employees/:id` | profile + role |
| GET | `/employees/:id/history` | all time (`days: null`) |
| PUT | `/employees/:id/name` | `{firstName, lastName}`, `If-Match`; the only way to change a name |
| POST | `/employees` | create account + profile; returns a one-time temporary password |
| POST | `/employees/:id/reset-password` | new temporary password; user must change it at next sign-in |
| GET / POST | `/skills` | list with usage / create |
| PATCH | `/skills/:id` | `{name?, isActive?}`; skills are turned off, never deleted |
| PUT | `/skills/order` | `[{id, sortOrder}]`, all or nothing |
| GET | `/settings` | |
| PUT | `/settings/history-days` | `{value}` 1–365, applies immediately |

Every write takes `If-Match` with the version from the last read. A stale version returns
`409 VERSION_CONFLICT`; the update and its history entry are committed in one transaction.
Errors use one envelope: `{"error": {"code", "message", "fields", "requestId"}}`.

## Tests

```bash
make test               # unit tests (no database)
make test-integration   # PostgreSQL 17 in Docker via testcontainers
```

Integration tests cover concurrent edits (one wins, the rest get 409), rollback when the
history insert fails, the history window following the admin setting without a restart,
photo cleanup on failure, and the database role being unable to rewrite history.

## Backend layout

Clean Architecture, one package per feature:

```
backend/
├── app/main.go               entry point
├── modules/
│   ├── entity/               domain types and business rules (no framework imports)
│   ├── auth/ profile/ skill/ history/ file/ setting/ employee/
│   │     handler.go → usecase.go → repository.go
│   └── server/               router; wires repository → use case → handler
└── pkg/                      database (migrations, SQL, sqlc), middleware, security, storage, imaging
```

## License

[MIT](LICENSE)
