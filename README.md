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
| GET  | `/profile` | `ETag: "<version>"` |
| PUT  | `/profile/contact/:field` | `mobile` \| `email` \| `address`, body `{value}`, header `If-Match` |
| PUT  | `/profile/photo` | multipart `file` (JPEG/PNG/WebP ≤ 5 MB), `If-Match` |
| PUT  | `/profile/skills` | `{skillIds: number[]}`, `If-Match` |
| GET  | `/profile/history` | `?field=&page=&pageSize=` |
| GET  | `/skills` | active skill catalog |
| GET  | `/files/photos/:name` | |
| GET  | `/healthz` | |

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
│   ├── auth/ profile/ skill/ history/ file/
│   │     handler.go → usecase.go → repository.go
│   └── server/               router; wires repository → use case → handler
└── pkg/                      database (migrations, SQL, sqlc), middleware, security, storage, imaging
```

## License

[MIT](LICENSE)
