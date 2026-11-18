# Employee Profile Portal

Employees view and update their own profile: photo, contact details (one field at a
time), and skills, with a change history for the last N days. Admins manage the skill
catalog and the history window.

**Backend:** Go 1.27 · Gin · PostgreSQL 17 · GORM (pgx driver) · golang-migrate
**Frontend:** React 19 · Vite · TypeScript · TanStack Query · React Router · Tailwind CSS · Radix UI

Full design and trade-offs: [docs/DESIGN.md](docs/DESIGN.md)

## Run it

Requires Docker.

```bash
make up          # creates .env with a random JWT secret, then starts postgres → migrations → API → web
open http://localhost:5173
```

For frontend work with hot reload: `make up` once, then `make web` (Vite proxies `/api` to :8080;
stop the `web` container first, or Vite picks the next free port).

| Account     | Password   | What it shows                                              |
|-------------|------------|------------------------------------------------------------|
| `EMP-00001` | `demo1234` | Complete profile; history inside and outside the 7-day window |
| `EMP-00002` | `demo1234` | Empty profile (first-time experience)                      |
| `ADMIN-001` | `demo1234` | Admin role: lands on `/admin`, no Profile menu             |

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
| GET | `/employees/:id/history` | everything kept (`days: null`) |
| PUT | `/employees/:id/name` | `{firstName, lastName}`, `If-Match`; the only way to change a name |
| POST | `/employees` | create account + profile; returns a one-time temporary password |
| POST | `/employees/:id/reset-password` | new temporary password; user must change it at next sign-in |
| GET / POST | `/skills` | list with usage / create |
| PATCH | `/skills/:id` | `{name?, isActive?}`; skills are turned off, never deleted |
| PUT | `/skills/order` | `[{id, sortOrder}]`, all or nothing |
| GET | `/settings` | `historyDays` (window shown) and `maxHistoryDays` (retention) |
| PUT | `/settings/history-days` | `{value}` from 1 to the retention, applies immediately |

Every write takes `If-Match` with the version from the last read. A stale version returns
`409 VERSION_CONFLICT`; the update and its history entry are committed in one transaction.
Errors use one envelope: `{"error": {"code", "message", "fields", "requestId"}}`.

**History retention.** Two numbers: employees *see* the last `historyDays` (admin setting,
default 7), and the server *keeps* `HISTORY_RETENTION_DAYS` (env, default 90). A job in the
API (`RETENTION_INTERVAL`, default daily) deletes older changes as the separate `retention`
database role; the API role can only append to the log. Replaced photos are kept so the
history can show them: beyond the retention, the 3 newest photos per employee are still
kept, and a photo file is deleted only when no kept change or profile uses it.

## Tests

```bash
make test               # Go unit tests (no database)
make test-integration   # Go integration tests: PostgreSQL 17 in Docker via testcontainers
make test-web           # frontend typecheck + Vitest / Testing Library
make e2e                # Playwright against the full stack: desktop, iPhone 13 and iPad
```

Integration tests cover concurrent edits (one wins, the rest get 409), rollback when the
history insert fails, the history window following the admin setting without a restart,
photo cleanup on failure, and the database role being unable to rewrite history.
End-to-end tests create a fresh employee per test, then edit contact details, skills and
the photo, check the history and the admin flows, and assert there is no horizontal scroll.

## Frontend layout

```
frontend/src/
├── api/              fetch client (error envelope, If-Match) and API types
├── components/       app shell + a small UI kit (Radix + Tailwind)
├── features/
│   ├── auth/         sign-in, forced password change, route guards
│   ├── profile/      photo, per-field contact edit, skills picker, checklist, recent changes
│   ├── history/      timeline shared by the employee and admin views
│   └── admin/        dashboard, employees, skills, settings
└── lib/              formatting, media query, debounce
```

Phones get a single column and bottom sheets for editing; tablets and desktops get inline
editing and popovers. Every edit sends `If-Match`; a 409 reloads the profile and says why.

## Backend layout

Clean Architecture, packaged by feature. Each feature splits into one package per layer:

```
backend/
├── app/main.go               entry point
├── modules/
│   ├── shared/               Actor, Role, domain errors, paging (imports nothing from the project)
│   ├── auth/ profile/ skill/ history/ setting/ employee/
│   │   ├── domain/           business rules and types (no framework imports)
│   │   ├── model/            GORM structs for this feature's tables
│   │   ├── repo/             GORM queries; model ↔ domain
│   │   ├── service/          use cases + transactions; interface.go = ports it needs (Repository …)
│   │   └── api/              Gin handlers + DTOs; interface.go = the Service it calls
│   ├── file/api/             streams stored photos (no other layers)
│   └── server/               wire.go builds repo → service → api; routes.go is the one route table
└── pkg/                      database (GORM, migrations), middleware, httputil, security, storage, imaging
```

Dependencies point inward: `api → service → domain`, `repo → service (ports), model, domain`.
A feature's `domain` may import another feature's `domain` (profile uses skill and history),
but `model`, `repo`, `service` and `api` never import another feature's.

## License

[MIT](LICENSE)
