# Employee Profile Portal — Design Document

> Full Stack Developer Assignment · Frontend: React (Vite) · Backend: Go (Gin) · DB: PostgreSQL 17 (pgx + sqlc)

---

## 1. Scope

### 1.1 Functional Requirements (จากโจทย์)
| # | Requirement | หมายเหตุการออกแบบ |
|---|---|---|
| F1 | ผู้ใช้ดูโปรไฟล์ตัวเอง: ชื่อ–นามสกุล, รูป, ข้อมูลติดต่อ, Skills | `GET /api/v1/profile` |
| F2 | ชื่อ–นามสกุล **read-only** | ไม่มี endpoint สำหรับผู้ใช้แก้ → บังคับที่ server ไม่ใช่แค่ UI |
| F3 | อัปโหลดรูปโปรไฟล์ | multipart, ตรวจ magic bytes, resize |
| F4 | แก้ข้อมูลติดต่อ **แยกทีละ field** (Mobile, Email, Address) | 1 field = 1 request, มี version ป้องกันการเขียนทับ |
| F5 | เลือก Skills ได้หลายตัว จากรายการที่ **configurable** | รายการอยู่ในตาราง `skills` (seed จาก config) |
| F6 | ดูประวัติการแก้ไขย้อนหลัง **7 วัน — configurable** | ค่าใน `app_settings.history_days` (fallback: env `HISTORY_DAYS`) |

### 1.2 Extra Features (ผู้สมัครเพิ่มเอง)
| # | Feature | ลำดับ |
|---|---|---|
| X1 | Responsive (mobile-first) ทุกหน้า | หลัก |
| X2 | First-time experience: empty state + checklist ความครบของโปรไฟล์ (ไม่มีหน้า onboarding แยก) | หลัก |
| X3 | Admin: จัดการ Skills + ตั้งค่า history days (ทำให้ "configurable" ปรับได้จากหน้าจอจริง) | Tier 1 |
| X4 | Admin: รายชื่อพนักงาน, ดูโปรไฟล์/ history, แก้ชื่อ–นามสกุล | Tier 2 |
| X5 | Admin: สร้างบัญชี, รีเซ็ตรหัส, บังคับเปลี่ยนรหัสตอน login ครั้งแรก | Tier 3 |

### 1.3 Confirmed with Team
- DB: **PostgreSQL** (ทีมกำลังย้ายจาก MySQL Stored Procedure → ไม่บังคับใช้ SP)
- Go framework/library: เลือกเองได้ → ใช้ **Gin**
- ส่วนอื่นเปิดกว้าง · GitHub repo เป็น **public**

### 1.4 Assumptions
- มีระบบ Login แบบง่าย (seed users) — login ด้วย `employee_code` ไม่ใช้อีเมลติดต่อ เพราะอีเมลแก้ไขได้
- มี 2 role: `employee`, `admin` — บัญชีถูกสร้างโดย HR/Admin (ไม่มี sign-up)
- "7 วัน" = ช่วงที่**แสดง** ไม่ใช่ช่วงที่**เก็บ** → เก็บ log ทั้งหมด, purge เป็น optional job
- Change history อยู่ใน PostgreSQL ตัวเดียวกัน เพื่อให้ update + log อยู่ใน transaction เดียว (ไม่แยก Mongo — ดู §12)
- **Business logic อยู่ใน Go (service layer)** · SQL เป็น query ธรรมดาที่ generate เป็นโค้ด Go ด้วย sqlc · transaction ควบคุมจาก service

---

## 2. Architecture

```
┌──────────────────────┐        HTTPS / JSON         ┌─────────────────────────────┐
│  React SPA (Vite+TS) │ ──────────────────────────▶ │  Go API (Gin)               │
│  TanStack Query      │ ◀────────────────────────── │  handler → service → repo   │
└──────────────────────┘   httpOnly JWT cookie       └──────┬───────────────┬──────┘
                                                             │ pgx (sqlc)    │ write/read file
                                                             ▼               ▼
                                                    ┌────────────────┐ ┌──────────────┐
                                                    │ PostgreSQL 17  │ │ File storage │
                                                    │                │ │ (local vol)  │
                                                    └────────────────┘ └──────────────┘
                 ทั้งหมดรันด้วย docker compose (postgres, migrate, api, web)
```

### 2.1 Tech Stack
| Layer | เลือกใช้ |
|---|---|
| Frontend | React 19 + TypeScript 5.9, Vite 8, React Router 8, TanStack Query 5, React Hook Form + Zod 4, Tailwind 4 + Radix UI (component เขียนเองแบบ shadcn), sonner (toast), Vitest + Testing Library, Playwright |
| HTTP | `gin-gonic/gin`, `gin-contrib/cors` |
| DB driver | `jackc/pgx/v5` (`pgxpool`) |
| Query | `sqlc` — เขียน SQL ในไฟล์ `.sql` → generate ฟังก์ชัน Go แบบ type-safe (ตรวจ query ตั้งแต่ตอน generate) |
| Migration | `golang-migrate/migrate` |
| Auth | `golang-jwt/jwt/v5`, `golang.org/x/crypto/bcrypt` |
| Config | `caarlos0/env` (อ่านจาก env / `.env`) |
| Validation | `go-playground/validator` (มากับ Gin binding) |
| Image | `disintegration/imaging` (resize → JPEG) |
| Rate limit | `golang.org/x/time/rate` |
| Logging | `log/slog` (JSON) + request-id middleware |
| Test | `testify`, `httptest`, `testcontainers-go` (Postgres module), Vitest + RTL, Playwright |

---

## 3. Repository Structure

Backend ใช้ **Clean Architecture แบบ Package by Feature**: 1 module = 1 package แยก layer ด้วยไฟล์ (`handler.go` → `usecase.go` → `repository.go`) และใช้ `entity` กับ `pkg/` ร่วมกัน

```
employee-portal/
├── backend/
│   ├── app/main.go                     # config → slog → pgxpool → server.Run (graceful shutdown)
│   ├── configs/config.go               # env → struct (JWT_SECRET required, ≥ 32 chars)
│   ├── modules/
│   │   ├── server/
│   │   │   ├── server.go               # gin.Engine, global middleware, Run/Shutdown
│   │   │   └── routes.go               # composition root: repo → usecase → handler → route group
│   │   ├── entity/                     # domain ล้วน: ไม่มี tag, ไม่ import gin/pgx
│   │   │   ├── employee.go             # Employee, Profile, ChangeContact()
│   │   │   ├── contact.go              # ContactField: Parse / Normalize / Validate
│   │   │   ├── changelog.go · skill.go · user.go · actor.go
│   │   │   ├── errors.go               # ErrNotFound, ErrConflict, ErrForbidden, ErrUnauthorized, ValidationError
│   │   │   └── *_test.go
│   │   ├── auth/                       # handler.go · usecase.go · usecase_test.go · repository.go
│   │   ├── profile/                    # handler.go · usecase.go · usecase_test.go · repository.go (รวม PUT /profile/skills)
│   │   ├── skill/                      # GET /skills (catalog สำหรับ picker)
│   │   ├── history/                    # GET /profile/history (อ่านอย่างเดียว — การเขียน log อยู่ใน transaction ของ profile)
│   │   ├── file/                       # GET /files/photos/:name (handler อย่างเดียว — stream ไฟล์)
│   │   ├── setting/                    # Admin → Settings (history days)
│   │   ├── employee/                   # Admin → Employees (list, search, detail) + Dashboard stats
│   ├── pkg/
│   │   ├── database/
│   │   │   ├── postgres.go             # pgxpool + ping
│   │   │   ├── tx.go                   # WithTx(ctx, pool, fn(*sqlc.Queries))
│   │   │   ├── errors.go               # pg error → entity error
│   │   │   ├── migrations/             # golang-migrate: init, grants, seed_demo
│   │   │   ├── queries/*.sql           # input ของ sqlc
│   │   │   └── sqlc/                   # generated (ห้ามแก้มือ)
│   │   ├── middleware/                 # requestid, logger, recover, auth, rbac, ratelimit, headers, timeout
│   │   ├── httputil/                   # error envelope (WriteError), ETag / If-Match
│   │   ├── security/                   # jwt.go, password.go (bcrypt)
│   │   ├── storage/                    # Local: ชื่อ UUID, เขียน temp → rename, Open/Delete ตรวจชื่อ
│   │   ├── imaging/                    # ตรวจ magic bytes, resize ≤ 512, re-encode JPEG
│   ├── db-init/01_roles.sql            # สร้าง role app (LOGIN) ตอน postgres init ครั้งแรก — local dev
│   ├── tests/integration/              # testcontainers Postgres 17 (build tag: integration)
│   ├── sqlc.yaml
│   ├── Dockerfile                      # distroless, nonroot
│   └── go.mod
├── frontend/
│   ├── src/
│   │   ├── api/                        # fetch client + types
│   │   ├── features/auth/
│   │   ├── features/profile/           # ProfileHeader, ProfileChecklist, ContactField, SkillsPicker, RecentChangesCard
│   │   ├── features/history/           # HistoryTimeline
│   │   ├── features/admin/             # AdminShell, Skills, Settings, Employees, Dashboard
│   │   ├── components/ui/
│   │   └── lib/
│   └── Dockerfile
├── docs/                               # DESIGN.md, screenshots
├── docker-compose.yml
├── Makefile                            # up, down, migrate, sqlc, seed, test, lint
├── .env.example
├── .github/workflows/ci.yml
├── LICENSE
└── README.md
```

---

## 4. Database Schema (PostgreSQL 17)

```sql
CREATE TABLE users (
  id                    BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  employee_code         TEXT        NOT NULL UNIQUE CHECK (length(employee_code) BETWEEN 3 AND 20),
  password_hash         TEXT        NOT NULL,
  role                  TEXT        NOT NULL DEFAULT 'employee' CHECK (role IN ('employee','admin')),
  is_active             BOOLEAN     NOT NULL DEFAULT TRUE,
  must_change_password  BOOLEAN     NOT NULL DEFAULT FALSE,           -- X5 (optional)
  created_at            TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE app_settings (
  key         TEXT PRIMARY KEY,
  value       TEXT NOT NULL,
  updated_by  BIGINT REFERENCES users(id),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- seed: ('history_days', '7')

CREATE TABLE employees (
  id             BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  user_id        BIGINT      NOT NULL UNIQUE REFERENCES users(id),
  first_name     TEXT        NOT NULL CHECK (length(first_name) BETWEEN 1 AND 100),
  last_name      TEXT        NOT NULL CHECK (length(last_name)  BETWEEN 1 AND 100),
  photo_path     TEXT,
  mobile_phone   TEXT        CHECK (mobile_phone  ~ '^\+?[0-9]{9,15}$'),
  contact_email  TEXT        CHECK (length(contact_email) <= 254),
  address        TEXT        CHECK (length(address) BETWEEN 1 AND 500),
  version        INTEGER     NOT NULL DEFAULT 1,                       -- optimistic locking
  updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE skills (
  id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  name        TEXT    NOT NULL CHECK (length(name) BETWEEN 1 AND 50),
  is_active   BOOLEAN NOT NULL DEFAULT TRUE,
  sort_order  INTEGER NOT NULL DEFAULT 0
);
CREATE UNIQUE INDEX uq_skills_name ON skills (lower(name));         -- "go" ซ้ำกับ "Go" ไม่ได้

CREATE TABLE employee_skills (
  employee_id BIGINT NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
  skill_id    BIGINT NOT NULL REFERENCES skills(id),
  PRIMARY KEY (employee_id, skill_id)
);

CREATE TABLE profile_change_logs (
  id           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  employee_id  BIGINT      NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
  field        TEXT        NOT NULL CHECK (field IN ('first_name','last_name','mobile_phone',
                                                     'contact_email','address','photo','skills')),
  old_value    JSONB,                  -- string หรือ array (skills) ใช้ชนิดเดียวกันได้
  new_value    JSONB,
  changed_by   BIGINT      NOT NULL REFERENCES users(id),
  changed_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_log_emp_time ON profile_change_logs (employee_id, changed_at DESC);
```

- `TEXT` + `CHECK` แทน `VARCHAR(n)` / `ENUM` — เพิ่มค่าใหม่ได้ด้วย migration ธรรมดา
- `old_value` / `new_value` เป็น **JSONB** — เบอร์โทรเก็บเป็น `"0812345678"`, skills เก็บเป็น `["Go","React"]` ใน column เดียวกัน FE ได้ type ตรง ๆ
- `TIMESTAMPTZ` ทุกที่ (เก็บ UTC แสดงตาม timezone ผู้ใช้)

**ER สรุป:** `users 1—1 employees 1—N employee_skills N—1 skills`, `employees 1—N profile_change_logs`, `app_settings` (key-value)

---

## 5. Data Access (sqlc + transaction ใน service)

### 5.1 แนวทาง
- **Logic อยู่ใน Go service**: validation, diff สำหรับ history, ตรวจสิทธิ์, จัดการ transaction
- **SQL อยู่ในไฟล์ `.sql`** → `sqlc generate` ได้ฟังก์ชัน Go ที่ type-safe; ตอน walkthrough อ่าน SQL จริงได้ทุกบรรทัด
- **ทุก use case ที่เขียนข้อมูล** = 1 transaction: lock แถว → update (เช็ก version) → insert log → commit

### 5.2 Queries หลัก
| Query (sqlc name) | ชนิด | ใช้ที่ |
|---|---|---|
| `GetUserByCode` | `:one` | Login |
| `GetSetting` / `UpsertSetting` | `:one` / `:exec` | F6, Admin |
| `GetProfile` | `:one` | F1 |
| `GetProfileForUpdate` | `:one` (`FOR UPDATE`) | อ่านค่าเดิม + lock แถวก่อนแก้ |
| `ListEmployeeSkills` | `:many` | F1, F5 (ชุดเดิม) |
| `ListActiveSkills` | `:many` | F5 (picker) |
| `GetSkillsByIDs` | `:many` | F5 — คืนทั้ง active/inactive เพื่อแยก "ไม่มีจริง" กับ "ถูกปิด" |
| `UpdateContactField` | `:one` (`RETURNING version`) | F4 |
| `UpdatePhoto` | `:one` | F3 |
| `BumpVersion` | `:one` | F5 (skills ไม่ได้อยู่ในตาราง employees) |
| `DeleteEmployeeSkills` / `InsertEmployeeSkills` | `:exec` | F5 |
| `InsertChangeLog` | `:exec` | ทุกการแก้ไข |
| `ListHistory` / `HasOlderHistory` | `:many` / `:one` | F6 |
| `Admin*` (list/upsert/toggle/reorder skills, list employees, update name, stats) | – | X3–X4 |

### 5.3 ตัวอย่าง SQL
```sql
-- name: GetProfileForUpdate :one
SELECT id, first_name, last_name, photo_path, mobile_phone, contact_email, address, version
FROM employees
WHERE id = $1
FOR UPDATE;

-- name: UpdateContactField :one
-- column ถูกเลือกจาก whitelist ใน Go (contactRules) แล้วส่งชื่อเข้ามาเป็นค่า ไม่ใช่ประกอบ SQL เอง
UPDATE employees SET
  mobile_phone  = CASE WHEN sqlc.arg(field)::text = 'mobile_phone'  THEN sqlc.narg(value)::text ELSE mobile_phone  END,
  contact_email = CASE WHEN sqlc.arg(field)::text = 'contact_email' THEN sqlc.narg(value)::text ELSE contact_email END,
  address       = CASE WHEN sqlc.arg(field)::text = 'address'       THEN sqlc.narg(value)::text ELSE address       END,
  version       = version + 1,
  updated_at    = now()
WHERE id = sqlc.arg(id) AND version = sqlc.arg(version)
RETURNING version;

-- name: InsertEmployeeSkills :exec
INSERT INTO employee_skills (employee_id, skill_id)
SELECT sqlc.arg(employee_id), unnest(sqlc.arg(skill_ids)::bigint[]);

-- name: ListHistory :many
-- days NULL = no time window (admin). field NULL = all fields.
-- Fetch page_limit = pageSize + 1 rows to know whether a next page exists.
SELECT l.id, l.field, l.old_value, l.new_value, l.changed_at, l.changed_by, u.role AS changed_by_role
FROM profile_change_logs l
JOIN users u ON u.id = l.changed_by
WHERE l.employee_id = sqlc.arg(employee_id)
  AND (sqlc.narg(days)::int IS NULL OR l.changed_at >= now() - make_interval(days => sqlc.narg(days)::int))
  AND (sqlc.narg(field)::text IS NULL OR l.field = sqlc.narg(field)::text)
ORDER BY l.changed_at DESC, l.id DESC
LIMIT sqlc.arg(page_limit)::int OFFSET sqlc.arg(page_offset)::int;

-- name: HasOlderHistory :one
SELECT EXISTS (
  SELECT 1 FROM profile_change_logs
  WHERE employee_id = sqlc.arg(employee_id)
    AND changed_at < now() - make_interval(days => sqlc.arg(days)::int)
);
```
- `days` เป็น nullable → `/admin/employees/:id/history` ใช้ query เดียวกันได้โดยไม่จำกัดช่วงเวลา
- `hasNext` ได้จากการดึงเกินมา 1 แถว (ไม่ต้อง `COUNT(*)`) · index `(employee_id, changed_at DESC)` รองรับทั้ง filter และ sort

### 5.4 Transaction (ตัวอย่าง: แก้ contact)
Use case ไม่รู้จัก sqlc/pgx: เรียก `WithTx` ผ่าน interface `Repository` ที่ use case ประกาศเอง และกฎธุรกิจอยู่ใน `entity.Employee.ChangeContact`
```go
// modules/profile/usecase.go
type Repository interface {
    GetProfile(ctx context.Context, employeeID int64) (entity.Profile, error)
    WithTx(ctx context.Context, fn func(tx TxRepository) error) error   // commit ถ้า nil, ไม่งั้น rollback
}
type TxRepository interface {
    GetForUpdate(ctx context.Context, employeeID int64) (entity.Employee, error)
    SaveContact(ctx context.Context, e entity.Employee, f entity.ContactField) (int32, error)
    AppendChange(ctx context.Context, employeeID int64, c entity.Change, by entity.Actor) error
}

func (u *usecase) UpdateContact(ctx context.Context, in UpdateContactInput) (int32, error) {
    field, err := entity.ParseContactField(in.Field)          // whitelist → 400 ถ้าไม่รู้จัก
    if err != nil {
        return 0, err
    }
    var newVersion int32
    err = u.repo.WithTx(ctx, func(tx TxRepository) error {
        emp, err := tx.GetForUpdate(ctx, in.Actor.EmployeeID)  // SELECT ... FOR UPDATE
        if err != nil {
            return err                                          // ไม่มีแถว → ErrNotFound
        }
        change, err := emp.ChangeContact(field, in.Value, in.Version) // normalize, validate, เช็ก version
        if err != nil {
            return err                                          // ErrValidation / ErrConflict
        }
        if newVersion, err = tx.SaveContact(ctx, emp, field); err != nil {
            return err
        }
        if change == nil {
            return nil                                          // ค่าเดิม → version +1 แต่ไม่ log
        }
        return tx.AppendChange(ctx, emp.ID, *change, in.Actor)
    })
    return newVersion, err
}
```
- `repository.go` แปลง sqlc row ↔ entity, map ชื่อ column จาก `contactColumn` (whitelist) และแปลง pg error เป็น entity error ผ่าน `database.MapError`
- ตรวจ version หลัง `FOR UPDATE` → แยก 404 (ไม่มีแถว) กับ 409 (version ไม่ตรง) ได้ชัด
- Unit test ของ use case ใช้ fake repository (in-memory) ส่วน transaction, lock และ rollback จริงทดสอบใน `tests/integration`
- `SetSkills`: ลำดับเดียวกัน แต่ใช้ `ReplaceSkills` + `BumpVersion` แทน `SaveContact` (รายละเอียด §8)

### 5.5 Error Mapping (pg → domain → HTTP)
| PostgreSQL | ความหมาย | Domain error | HTTP |
|---|---|---|---|
| `pgx.ErrNoRows` | ไม่พบข้อมูล | `ErrNotFound` | 404 |
| `23505` unique_violation | ชื่อ skill ซ้ำ | `ErrDuplicate` | 409 `ALREADY_EXISTS` |
| `23514` check_violation | ละเมิด CHECK (เช่น เบอร์ผิดรูปแบบ) | `ErrValidation` | 400 |
| `23503` foreign_key_violation | อ้างอิง id ที่ไม่มี | `ErrValidation` | 400 |
| version ไม่ตรง (ตรวจใน service) | มีคนแก้ก่อน | `ErrConflict` | 409 |
| role ไม่ใช่ admin (ตรวจใน service) | ไม่มีสิทธิ์ | `ErrForbidden` | 403 |
| อื่น ๆ | – | wrap เป็น internal | 500 |

---

## 6. REST API (`/api/v1`)

| Method | Path | Auth | Body / Query | Response |
|---|---|---|---|---|
| POST | `/auth/login` | – | `{employeeCode, password}` | 204 + Set-Cookie |
| POST | `/auth/logout` | ✔ | – | 204 |
| GET | `/auth/me` | ✔ | – | `{userId, employeeId, employeeCode, role, mustChangePassword}` (ใช้ตอนแอปโหลด) |
| POST | `/auth/change-password` | ✔ | `{currentPassword, newPassword}` | 204 + Set-Cookie ใหม่ (ไม่มี flag `mcp`) · รหัสเดิมผิด → **400** (ไม่ใช่ 401 เพื่อไม่ให้ FE เด้งออก) |
| GET | `/profile` | ✔ | – | `Profile` + header `ETag: "<version>"` |
| PUT | `/profile/contact/:field` | ✔ | `field` = `mobile` \| `email` \| `address` · body `{value}` + `If-Match` | `{version}` · field อื่น → 400 |
| PUT | `/profile/photo` | ✔ | multipart `file` + `If-Match` | `{version, photoUrl}` |
| GET | `/skills` | ✔ | – | `[{id, name}]` |
| PUT | `/profile/skills` | ✔ | `{skillIds: number[]}` + `If-Match` | `{version}` |
| GET | `/profile/history` | ✔ | `?field=&page=&pageSize=` · `field` = ชื่อเดียวกับ `items[].field` (`mobile_phone`, `contact_email`, `address`, `photo`, `skills`, `first_name`, `last_name`) · `pageSize` 1–100 (default 20) | `{days, hasOlder, items: [...], page}` |
| GET | `/files/photos/:name` | ✔ | – | image |
| GET | `/healthz` | – | – | `{status, db}` |

### Admin (`RequireRole("admin")`)
| Method | Path | Body / Query | Response |
|---|---|---|---|
| GET | `/admin/stats` | – | `{employees, completeProfiles, topSkills: [{id, name, usedBy}] (5), recentChanges: [{employeeId, employeeCode, name, field, changedAt, changedByRole}] (5)}` · "ครบ" = รูป + contact ครบ 3 + skill ≥ 1 (ตรงกับ checklist) |
| GET | `/admin/skills` | – | `[{id, name, isActive, sortOrder, usedBy}]` (รวม inactive) |
| POST | `/admin/skills` | `{name}` (trim + ยุบช่องว่าง, 1–50 ตัว) | 201 `{id}` · ต่อท้ายรายการ · 409 `ALREADY_EXISTS` ถ้าชื่อซ้ำ (ไม่สนตัวพิมพ์) |
| PATCH | `/admin/skills/:id` | `{name?, isActive?}` (ต้องมีอย่างน้อย 1) | 204 · 404 id ไม่มี · 409 ชื่อซ้ำ |
| PUT | `/admin/skills/order` | `[{id, sortOrder}]` | 204 · id ไม่รู้จักแม้ตัวเดียว → 404 และไม่มีอะไรเปลี่ยน (transaction) |
| GET | `/admin/settings` | – | `{historyDays, updatedBy: {id, employeeCode} \| null, updatedAt \| null}` |
| PUT | `/admin/settings/history-days` | `{value}` (จำนวนเต็ม 1–365) | 204 · มีผลกับ `/profile/history` ทันที |

module: จัดการ skill อยู่ใน `skill/` (เจ้าของ catalog) · ตั้งค่าอยู่ใน `setting/` · รายชื่อ/รายละเอียด/stats อยู่ใน `employee/` · history ของ admin อยู่ใน `history/` · แก้ชื่ออยู่ใน `profile/` (ใช้ lock + version + log เดียวกับการแก้โปรไฟล์) · ทุก admin use case เรียก `entity.RequireAdmin(actor)` ซ้ำหลัง `RequireRole` ใน router
| GET | `/admin/employees` | `?q=&page=&pageSize=` · `q` ค้นรหัสพนักงานหรือ "ชื่อ นามสกุล" (ไม่สนตัวพิมพ์, `%` `_` ที่พิมพ์มาเป็นตัวอักษรธรรมดา) | `{items: [{id, employeeCode, firstName, lastName, photoUrl, role, isActive}], page: {page, pageSize, hasNext}, total}` |
| GET | `/admin/employees/:id` | – | `Profile` + `role`, `isActive`, skills มี `isActive` · `ETag` สำหรับแก้ชื่อ |
| GET | `/admin/employees/:id/history` | `?field=&page=&pageSize=` | รูปเดียวกับ `/profile/history` แต่ `days: null` = ทั้งหมด (ไม่จำกัดช่วงเวลา) · 404 ถ้าไม่มีพนักงาน |
| PUT | `/admin/employees/:id/name` | `{firstName, lastName}` (trim, 1–100) + `If-Match` | `{version}` · log `first_name` / `last_name` แยกแถว, `changed_by` = admin |
| POST | `/admin/employees` | `{employeeCode, firstName, lastName, role?}` | 201 `{id, employeeCode, temporaryPassword}` · สร้าง user + employee ใน transaction เดียว · รหัสชั่วคราวแสดงครั้งเดียว · 409 ถ้ารหัสพนักงานซ้ำ |
| POST | `/admin/employees/:id/reset-password` | – | `{temporaryPassword}` · บังคับเปลี่ยนรหัสตอน login ครั้งถัดไป |

### 6.1 ตัวอย่าง Response
```json
// GET /profile
{
  "id": 1,
  "firstName": "Somchai",
  "lastName": "Jaidee",
  "photoUrl": "/api/v1/files/photos/9f2c...jpg",
  "contact": { "mobilePhone": "0812345678", "email": "somchai@example.com", "address": "..." },
  "skills": [{ "id": 1, "name": "Go" }, { "id": 2, "name": "React" }],
  "version": 7
}

// GET /profile/history
{
  "days": 7,
  "hasOlder": true,
  "items": [
    { "id": 31, "field": "skills", "oldValue": ["Go","React"], "newValue": ["Go","Flutter"],
      "changedAt": "2026-11-16T09:12:03.120Z", "changedBy": { "id": 5, "role": "employee" } },
    { "id": 30, "field": "mobile_phone", "oldValue": "0811111111", "newValue": "0812345678",
      "changedAt": "2026-11-15T14:01:44.002Z", "changedBy": { "id": 5, "role": "employee" } }
  ],
  "page": { "page": 1, "pageSize": 20, "hasNext": false }
}
```

### 6.2 Error Envelope (ทุก endpoint)
```json
{ "error": { "code": "VERSION_CONFLICT", "message": "Profile was updated elsewhere. Please reload.",
             "fields": null, "requestId": "01JD..." } }
```
| code | HTTP |
|---|---|
| `VALIDATION_ERROR` (มี `fields`) | 400 |
| `UNAUTHORIZED` | 401 |
| `FORBIDDEN` | 403 |
| `PASSWORD_CHANGE_REQUIRED` | 403 |
| `NOT_FOUND` | 404 |
| `METHOD_NOT_ALLOWED` | 405 |
| `VERSION_CONFLICT` | 409 |
| `ALREADY_EXISTS` (เช่น ชื่อ skill ซ้ำ) | 409 |
| `PAYLOAD_TOO_LARGE` | 413 |
| `PRECONDITION_REQUIRED` (ไม่ส่ง `If-Match`) | 428 |
| `UNSUPPORTED_MEDIA_TYPE` | 415 |
| `RATE_LIMITED` | 429 |
| `INTERNAL_ERROR` | 500 |

### 6.3 Validation Rules (Go — ชั้นแรก)
| Field | Rule |
|---|---|
| mobile | trim, ตัด `-` / ช่องว่าง / วงเล็บ, `^\+?[0-9]{9,15}$` (กฎเดียวกับ CHECK ใน DB) |
| email | trim, lowercase, `net/mail` ต้องเป็น address เปล่า (ไม่รับ `Name <a@b>`), ≤ 254 |
| address | trim, 1–500 ตัวอักษร (นับเป็น rune — ภาษาไทยนับถูก) |
| contact ทุก field | กรอกค่าได้แต่ล้างเป็นค่าว่างไม่ได้ (ตั้งใจ — ลดความกำกวมของ "ลบ" กับ "ไม่ส่ง") |
| photo | ≤ 5 MB, magic bytes = jpeg/png/webp, decode ได้, resize ≤ 512×512 → JPEG q85 |
| skillIds | ต้องส่ง (`[]` = ล้างทั้งหมด), ไม่ซ้ำ, ≤ 20 รายการ, ทุก id ต้องมีอยู่ และต้อง active **ยกเว้น**ตัวที่พนักงานมีอยู่แล้ว (ตรวจกับ DB) |

DB มี `CHECK` constraint เป็นชั้นที่สอง กันข้อมูลเสียจากทุกช่องทาง

---

## 7. Backend Design (Go + Gin)

### 7.1 Request Flow
```
Gin router
  → middleware : RequestID → Recover → Logger(slog) → SecureHeaders → CORS → Timeout → BodyLimit
                 (ต่อ route) RateLimit → Auth(JWT cookie) → RequireProfile / RequireRole
  → handler.go    : bind JSON, อ่าน If-Match, เรียก Usecase, เขียน envelope ผ่าน httputil.WriteError
  → usecase.go    : ลำดับขั้นของ use case + transaction — ไม่รู้จัก Gin / sqlc / pgx
  → entity        : กฎธุรกิจ (normalize, validate, version, diff สำหรับ history)
  → repository.go : เรียก sqlc, แปลง row ↔ entity, map pg error → entity error
```

**Dependency rule** (ลูกศร = import):
```
handler.go ──▶ usecase.go ──▶ entity
                   ▲
repository.go ─────┘ (implements interface ที่ usecase.go ประกาศ) ──▶ pkg/database/sqlc
```
- `entity` ไม่ import อะไรในโปรเจกต์เลย
- handler / usecase / repository อยู่ใน package เดียวกัน (ชื่อสั้น: `profile.NewUsecase`, ไม่ต้อง alias) คุม layer ด้วยการให้แต่ละชั้นรู้จักแค่ interface ของชั้นถัดไป
- `modules/server/routes.go` เป็นที่เดียวที่ประกอบทุกชั้นเข้าด้วยกัน

### 7.2 Interfaces หลัก
แต่ละ module ประกาศ interface ที่ use case **ต้องการ** ไว้ใน `usecase.go` ของตัวเอง (Dependency Inversion) — ดูตัวอย่าง `profile.Repository` / `profile.TxRepository` ใน §5.4
```go
// modules/auth/usecase.go
type Repository interface {
    GetUserByCode(ctx context.Context, code string) (entity.User, error)
    GetUserByID(ctx context.Context, id int64) (entity.User, error)
}
type TokenIssuer interface {
    Issue(a entity.Actor) (token string, expiresAt time.Time, err error)
}

// pkg/storage (ขั้นถัดไป)
type FileStorage interface {
    Save(ctx context.Context, r io.Reader, ext string) (path string, err error)
    Delete(ctx context.Context, path string) error
    Open(ctx context.Context, path string) (io.ReadCloser, error)
}
```

### 7.3 Contact: endpoint เดียว + กฎอยู่ใน entity
```go
// modules/entity/contact.go
type ContactField string                        // "mobile" | "email" | "address"
func ParseContactField(s string) (ContactField, error)
func (f ContactField) Normalize(v string) string
func (f ContactField) Validate(v string) error
func (f ContactField) LogField() FieldName      // mobile → mobile_phone (ชื่อใน history)

// modules/profile/repository.go — ชื่อ column อยู่ใน adapter เท่านั้น
var contactColumn = map[entity.ContactField]string{
    entity.ContactMobile:  "mobile_phone",
    entity.ContactEmail:   "contact_email",
    entity.ContactAddress: "address",
}
```
- เพิ่ม field ใหม่ = เพิ่ม const + กฎใน `contact.go`, 1 แถวใน `contactColumn` (+ migration + CASE ใน query)
- Unit test เป็น table-driven ใน `entity/contact_test.go`
- ชื่อ column มาจาก map เท่านั้น — ไม่มีการประกอบ SQL จาก input ผู้ใช้

### 7.4 Transaction (pgx)
```go
// pkg/database/tx.go
func WithTx(ctx context.Context, pool *pgxpool.Pool, fn func(q *sqlc.Queries) error) error {
    tx, err := pool.Begin(ctx)
    if err != nil {
        return err
    }
    defer tx.Rollback(ctx)                 // no-op ถ้า commit แล้ว
    if err := fn(sqlc.New(tx)); err != nil {
        return err
    }
    return tx.Commit(ctx)
}
```
Repository ของแต่ละ module ห่อ `WithTx` เป็น `WithTx(ctx, fn func(TxRepository) error)` เพื่อไม่ให้ `*sqlc.Queries` หลุดออกไปถึง use case

### 7.5 Flow อัปโหลดรูป (ไฟล์ + DB ให้สอดคล้องกัน)
```
1. imaging.Normalize: อ่านไม่เกิน UPLOAD_MAX_BYTES → magic bytes (jpeg/png/webp)
   → DecodeConfig (≤ 40 MP กัน decompression bomb) → decode + หมุนตาม EXIF
   → Fit 512×512 (ไม่ขยายรูปเล็ก) → ปูพื้นขาว (PNG โปร่งใส) → JPEG q85 (ตัด metadata/GPS ทิ้ง)
2. storage.Save()                         → <uuid>.jpg (เขียน temp แล้ว rename)
3. tx: lock → emp.ChangePhoto(name, version) → SavePhoto → AppendChange
   ├─ error → storage.Delete(new)         (ไม่เหลือไฟล์กำพร้า) → return error
   └─ ok    → storage.Delete(old)         (best-effort, log ถ้าพลาด)
```
- Body limit แยก route: JSON 64 KB · `PUT /profile/photo` = `UPLOAD_MAX_BYTES` + 64 KB (หัว multipart) · rate limit 10 ครั้ง/นาที/ผู้ใช้
- เกินขนาด → 413 `PAYLOAD_TOO_LARGE` · ไม่ใช่ jpeg/png/webp (ดูจาก magic bytes ไม่ใช่นามสกุล) → 415 `UNSUPPORTED_MEDIA_TYPE` · ไฟล์เสีย → 400
- `GET /files/photos/:name` ต้อง login · ชื่อไฟล์ต้องตรง `<uuid>.jpg` ไม่งั้น 404 (กัน path traversal) · `Cache-Control: private, max-age=31536000, immutable` (ชื่อไม่ซ้ำต่อการอัปโหลด)
- history ของ photo เก็บชื่อไฟล์ (`old_value` / `new_value`)

### 7.6 Auth & RBAC
- Login: `GetUserByCode` → ตรวจ `is_active` → `bcrypt.CompareHashAndPassword` → ออก JWT (HS256, `sub=user_id`, `emp=employee_id`, `role`, exp 8 ชม.)
- Error ของ login ไม่บอกว่าผิดที่รหัสพนักงานหรือรหัสผ่าน (กัน user enumeration)
- เก็บใน cookie `HttpOnly; Secure (prod); SameSite=Strict; Path=/api`
- Middleware ใส่ `Actor{UserID, EmployeeID, Role}` ลง `gin.Context` → handler ใช้ `/profile/*` ผู้ใช้จึงแก้ได้เฉพาะโปรไฟล์ตัวเอง
- RBAC 3 ชั้น: (1) FE ซ่อนเมนู + route guard → (2) Gin `RequireRole("admin")` → (3) `AdminService` ตรวจ `actor.Role` ซ้ำ (ชั้น 2–3 คือชั้นที่เชื่อถือได้)
  ```go
  api := r.Group("/api/v1")
  profile := api.Group("/profile", mw.Auth())   // employee id มาจาก JWT เสมอ — URL ไม่มี id
  admin := api.Group("/admin", mw.Auth(), mw.RequireRole("admin"))
  ```
- Rate limit `/auth/login` 5 req/นาที/IP (ปรับได้ด้วย `LOGIN_RATE_LIMIT` — E2E ตั้งสูงเพราะทุก test login จาก IP เดียว), `/auth/change-password` 5 req/นาที/user, upload 10 req/นาที/user
- CORS: request ที่ `Origin` ตรงกับ `Host` (same-origin ผ่าน nginx / Vite proxy) ข้าม allowlist — browser ส่ง `Origin` แม้เป็น same-origin POST · cross-origin ยังต้องอยู่ใน `CORS_ORIGINS` · nginx / Vite ส่ง `Host` เดิม (มี port) ต่อให้ API
- *(X5)* บัญชีใหม่และบัญชีที่ถูก reset มี `must_change_password = TRUE` → login ใส่ claim `mcp` ใน JWT → middleware `RequirePasswordChanged` ทำให้ API ทุกตัวยกเว้น `/auth/*` ตอบ 403 `PASSWORD_CHANGE_REQUIRED` จนกว่าจะเรียก `POST /auth/change-password` (ได้ cookie ใหม่ที่ไม่มี `mcp`)
  - รหัสชั่วคราว 12 ตัวจาก `crypto/rand` (ไม่มี 0/O/1/l/I) · เก็บแค่ bcrypt hash · policy รหัสใหม่: ≥ 8 ตัว, ≤ 72 bytes, มีตัวอักษรและตัวเลข, ต้องไม่ซ้ำรหัสเดิม
  - ข้อจำกัด: JWT แบบ stateless → reset รหัสแล้ว session เดิมของผู้ใช้ยังใช้ได้จนหมดอายุ (≤ 8 ชม.) · ทางแก้ในอนาคต: `token_version` ใน users แล้วเช็กใน middleware

### 7.7 Config (env)
| Key | Default | ใช้ทำอะไร |
|---|---|---|
| `APP_ENV` | `development` | `production` → Gin release mode + cookie `Secure` |
| `APP_PORT` | 8080 | |
| `LOG_LEVEL` | `info` | |
| `DATABASE_URL` | – (**required**) | เชื่อมด้วย role `app` (compose ใส่ให้) |
| `JWT_SECRET` | – (**required**, ≥ 32 ตัวอักษร, ไม่มีค่า default) | `make env` สุ่มให้ |
| `JWT_TTL` | `8h` | |
| `HISTORY_DAYS` | `7` | **F6** — ค่า fallback ถ้าไม่มีใน `app_settings` (ค่าจริงแก้ได้จากหน้า Admin) |
| `UPLOAD_DIR` | `/data/uploads` | |
| `UPLOAD_MAX_BYTES` | `5242880` | |
| `CORS_ORIGINS` | `http://localhost:5173` | |
| `TRUSTED_PROXIES` | (ว่าง) | IP ของ reverse proxy (เช่น web container) เพื่อให้ rate limit ใช้ IP จริงของผู้ใช้ |

Migration ใช้ role `owner` ผ่าน `docker-compose.yml` / `Makefile` (ไม่อยู่ใน config ของ API)

### 7.8 DB Roles & Privilege
```sql
-- owner: เป็นเจ้าของ schema, ใช้ตอน migrate เท่านั้น
-- app:   role ที่ API ใช้
--   - migration 000002 สร้าง app แบบ NOLOGIN (ถ้ายังไม่มี) แล้ว GRANT → ใช้ได้กับ DB ใหม่ทุกที่ รวมถึง testcontainers
--   - รหัสผ่าน/LOGIN ตั้งนอก migration: db-init/01_roles.sql (local dev), secret manager (prod)
GRANT SELECT, INSERT, UPDATE, DELETE ON users, employees, skills, employee_skills, app_settings TO app;
GRANT SELECT, INSERT ON profile_change_logs TO app;          -- ไม่มี UPDATE/DELETE → log แก้ย้อนหลังไม่ได้
GRANT USAGE ON ALL SEQUENCES IN SCHEMA public TO app;
```
- History เป็นแบบ **append-only** ที่ระดับสิทธิ์ DB — แม้มีบั๊กในแอปก็แก้/ลบ log ไม่ได้
- งาน purge (ถ้าทำ) ใช้ role แยกที่มีสิทธิ์ `DELETE`

---

## 8. Configurable Skills
- **Source of truth:** ตาราง `skills`
- ค่าเริ่มต้นอยู่ใน **migration** `000003_seed_demo.up.sql` (ไม่มีไฟล์ config แยก) → `docker compose up` ครั้งเดียวได้ catalog พร้อมใช้
- เปลี่ยน catalog ภายหลัง: เพิ่ม migration ใหม่ (ห้ามแก้ไฟล์ที่ apply ไปแล้ว) หรือจัดการผ่านหน้า **Admin → Skills** (เพิ่ม, แก้ชื่อ, เปิด/ปิด, จัดลำดับ)
- **ไม่ลบ skill** — ปิดด้วย `is_active = FALSE`: หายจาก picker (`GET /skills`) แต่พนักงานที่มีอยู่แล้วยังเก็บไว้ได้ และ history ยังอ่านได้

```sql
-- 000003_seed_demo.up.sql
INSERT INTO skills (name, sort_order) VALUES
  ('Go', 1), ('React', 2), ('Flutter', 3), ('CI/CD', 4),
  ('Docker', 5), ('Kubernetes', 6), ('PostgreSQL', 7), ('TypeScript', 8);
```

**`PUT /profile/skills` (แทนทั้งชุดใน request เดียว)** — `profile.SetSkills` ใน 1 transaction:
```
ValidateSkillIDs (≤ 20, ไม่ซ้ำ, > 0 — ตัดทิ้งก่อนเปิด tx)
→ GetForUpdate (lock)
→ ListEmployeeSkills (ชุดเดิม) + GetSkillsByIDs (ชุดใหม่ รวม inactive)
→ emp.ChangeSkills(requested, current, selected, version)   ← กฎทั้งหมดอยู่ที่นี่
     ValidateSkillIDs → CheckVersion → ทุก id ต้องมีจริง · inactive ได้เฉพาะตัวที่มีอยู่แล้ว
     → คืน Change (log ชื่อ ["Go","Flutter"] → ["Go","React"]) หรือ nil ถ้าชุดเดิม (ไม่สนลำดับ)
→ ReplaceSkills (delete + insert) → BumpVersion → AppendChange ถ้ามี Change
```
- pattern เดียวกับ contact: use case = โหลด → `emp.ChangeXxx(...)` → บันทึก → log
- `[]` = ล้างทุก skill (ได้) · ไม่ส่ง `skillIds` = 400
- กฎทดสอบใน `entity/employee_test.go` แบบไม่ต้องมี DB

---

## 9. Frontend Design

### 9.1 Stack Rationale
| เลือก | เหตุผล |
|---|---|
| **Vite** (ไม่ใช่ CRA / Next.js) | dev server เร็ว + HMR (ช่วยตอน live debug), build เป็น static files, proxy `/api` → Go ในตัว (ไม่ต้องตั้ง CORS ตอน dev), ใช้ Vitest config เดียวกัน · ไม่ต้องการ SSR/SEO เพราะเป็น portal ภายในที่ต้อง login |
| **Tailwind CSS** | utility-first, mobile-first breakpoints ใน class (`md:`, `lg:`) |
| **Radix UI + Tailwind** (แนวเดียวกับ shadcn/ui แต่เขียนเองไม่ใช้ CLI) | component อยู่ในโปรเจกต์ (`components/ui/` — อธิบาย/แก้ได้ทุกบรรทัด), a11y (focus trap, aria) จาก Radix · Dialog ตัวเดียวเป็นทั้ง dialog (≥ 768px) และ bottom sheet (มือถือ) |
| **nginx** เสิร์ฟ SPA + proxy `/api` | origin เดียวกับ API → cookie `SameSite=Strict; Path=/api` ใช้ได้โดยไม่ต้องตั้ง CORS · `/assets/*` cache ถาวร (ชื่อไฟล์มี hash), `index.html` no-cache |

```ts
// vite.config.ts
export default defineConfig({
  plugins: [react()],
  server: { proxy: { '/api': 'http://localhost:8080' } },
})
```

### 9.2 Routes & Navigation
| Path | หน้า | Guard |
|---|---|---|
| `/login` | Login | public (ถ้า login แล้ว → หน้าแรกตาม role: employee `/profile`, admin `/admin`) |
| `/change-password` | บังคับเปลี่ยนรหัส *(X5)* | auth |
| `/profile` | My Profile — **หน้าหลักของ employee** | employee (admin → `/admin`) |
| `/history` | Change History (เต็ม) | employee (admin → `/admin`) |
| `/admin` | Dashboard | admin |
| `/admin/skills` | จัดการ Skills | admin |
| `/admin/settings` | ตั้งค่า | admin |
| `/admin/employees` | รายชื่อพนักงาน | admin |
| `/admin/employees/:id` | รายละเอียด + แก้ชื่อ + history | admin |

**Navbar:** employee เห็น `Profile` · admin เห็น `Admin` อย่างเดียว (admin ไม่มีโปรไฟล์ของตัวเองให้แก้ — ดู/แก้ของพนักงานผ่าน `/admin/employees`) · Logout
- **ไม่มีเมนู History บน navbar** — เข้า `/history` ผ่านลิงก์ "View all" ในการ์ด Recent changes เท่านั้น (pattern สรุป → รายละเอียด, history เป็นส่วนหนึ่งของโปรไฟล์)
- มือถือ: เมนูน้อยจนไม่ต้องใช้ hamburger (ทั้ง employee และ admin) · เมนูย่อยของ admin เป็นแถบเลื่อนแนวนอนแทน Sheet (กดน้อยกว่า 1 ครั้ง, ไม่ซ่อนเมนู)
- เปลี่ยนรหัส: ไอคอนกุญแจบน navbar → `/change-password` · บัญชีที่ใช้รหัสชั่วคราวถูกพาไปหน้านี้ก่อนทุกหน้า

### 9.3 App Boot & Auth Flow
```
เปิด URL ใดก็ได้
  → GET /api/v1/auth/me  (ระหว่างรอ: full-page spinner สั้น ๆ)
     ├─ 401 → /login?next=<path>   → login สำเร็จ → กลับ <path> (default: employee /profile, admin /admin)
     ├─ mustChangePassword → /change-password   (X5)
     └─ ok  → render route  (<RequireAuth>, <RequireRole role="admin">)

ระหว่างใช้งาน: response 401 ใด ๆ → ล้าง cache → /login + toast "Your session expired"
```

### 9.4 หน้า Login
- การ์ดกลางจอ (max-width ~360px, มือถือเต็มความกว้าง): โลโก้ + "Employee portal"
- ช่อง **Employee code** + **Password** (ปุ่มแสดงรหัส) + ปุ่ม **Sign in** (loading state ระหว่างส่ง)
- Error รวม: "Employee code or password is incorrect" · 429 → "Too many attempts. Try again in a minute."
- "Forgot password?" / "Need access?" → ติดต่อ HR (reset ทางอีเมลอยู่นอก scope)

### 9.5 หน้า Profile
**Desktop (≥ 1024px)** — 2 คอลัมน์
```
┌──────────────┬────────────────────────────────────────┐
│ [Photo 📷]    │ Contact                                │
│ Somchai J. 🔒 │  Mobile   081 234 5678       [✎ Edit] │  ← ContactField × 3
│ EMP-00123    │  Email    [input........] [Cancel][Save]│     (edit/save/cancel แยกกัน)
│              │  Address  99 Sukhumvit...    [✎ Edit] │
│ Skills       ├────────────────────────────────────────┤
│ (Go ✕)(React)│ Recent changes · last 7 days  View all→│  ← 3 รายการล่าสุด
│ [+ Add]      │  Skills +CI/CD −Flutter         2h ago │
│              │  Mobile 081 111 → 081 234    Yesterday │
└──────────────┴────────────────────────────────────────┘
```
**Mobile (< 768px)** — คอลัมน์เดียว: Photo + ชื่อ → Contact (แต่ละแถวกดแล้วเปิด **bottom Drawer**) → Skills → Recent changes

**First-time / empty state** (ไม่มีหน้า onboarding แยก — ผู้ใช้เข้าใช้งานได้ทันที)
- Banner "Welcome, {firstName}" + progress bar + checklist 4 ข้อ: Name ✓, Photo, Contact info, Skills — คำนวณฝั่ง FE จากข้อมูลโปรไฟล์ · ครบ 4/4 แล้ว banner หายเอง
- ช่องว่าง → "Not added yet" + ปุ่ม **Add** (แทน Edit)
- รูปว่าง → วงกลมเส้นประ "Add photo" (กดเพื่ออัปโหลด)
- การ์ด Recent changes ว่าง (ข้อความ dynamic จาก `days` ของ API — ห้าม hard-code "7"):

| สถานการณ์ | ข้อความ |
|---|---|
| ยังไม่เคยแก้อะไร (`hasOlder = false`) | "No changes yet" · "Edits to your photo, contact info, and skills from the last {days} days will appear here." |
| เคยแก้ แต่เกินช่วง (`hasOlder = true`) | "No changes in the last {days} days" |

### 9.6 หน้า History (`/history`)
- หัวข้อ "Change history" + "Showing changes from the last {days} days"
- Filter chips: All · Photo · Mobile · Email · Address · Skills
- Timeline จัดกลุ่มตามวัน (Today / Yesterday / วันที่) · แต่ละรายการ: ไอคอน field, `ค่าเดิม` (ขีดฆ่า) → ค่าใหม่, เวลา, ผู้แก้ (ไอคอนคน = ตัวเอง, ไอคอนโล่ = admin)
- Skills แสดงเป็น diff สี: `+CI/CD` (เขียว) `−Flutter` (แดง)
- Timeline แบบเดียวกันทุกขนาดจอ (การ์ดต่อวัน — อ่านง่ายกว่าตารางเมื่อค่าเป็น diff ของ skills) · chip filter เลื่อนแนวนอนบนมือถือ · โหลดเพิ่มด้วย "Load more" (`useInfiniteQuery`)
- `HistoryTimeline` ใช้ร่วมกับหน้า admin (`/admin/employees/:id` แสดง "all time" เพราะ API คืน `days: null`)
- Empty state ใช้ข้อความเดียวกับ §9.5

### 9.7 หน้า Admin
Layout: sidebar (Dashboard, Employees, Skills, Settings) บน ≥ 1024px · ต่ำกว่านั้นเป็นแถบเลื่อนแนวนอน · สร้างบัญชี / รีเซ็ตรหัสแสดงรหัสชั่วคราวครั้งเดียวพร้อมปุ่ม Copy

| หน้า | องค์ประกอบ |
|---|---|
| Dashboard | การ์ดตัวเลข 3–4 ใบ: จำนวนพนักงาน, % โปรไฟล์ครบ, top 5 skills, การแก้ไขล่าสุด |
| Skills | รายการ: ปุ่ม ↑↓ ทุกขนาดจอ (แทน drag — ใช้ keyboard/screen reader ได้, ส่งลำดับใหม่ทั้งชุดใน request เดียว), Name (แก้ inline ✓/✕), Used by, สวิตช์ Active, ค้นหา, ปุ่ม Add skill · หมายเหตุ "Inactive skills disappear from the picker, but employees who already have them keep them." |
| Settings | Change history: input จำนวนวัน + ปุ่มลัด 7/14/30/90, hint "1–365 days. Older records are kept, just hidden.", แสดงผู้แก้ล่าสุด, Save มีผลทันที |
| Employees | ตารางค้นหาชื่อ/รหัส + pagination (มือถือเป็นการ์ด) |
| Employee detail | ข้อมูลโปรไฟล์, ช่อง Legal name (badge "Admin only") + Save name, history ทั้งหมดของคนนั้น (ระบุผู้แก้) |

### 9.8 Responsive Guidelines
| Breakpoint | ความกว้าง | Layout |
|---|---|---|
| (base) | < 640px | มือถือ — คอลัมน์เดียว |
| `sm:` | ≥ 640px | มือถือแนวนอน |
| `md:` | ≥ 768px | แท็บเล็ต — เริ่ม 2 คอลัมน์ |
| `lg:` | ≥ 1024px | เดสก์ท็อป — layout เต็ม, admin sidebar |

| ส่วน | Desktop | Mobile |
|---|---|---|
| แก้ contact | inline edit | bottom Drawer (คีย์บอร์ดไม่บัง) |
| เลือก skills | Popover + Command search | Drawer เต็มจอ + search |
| อัปโหลดรูป | ปุ่ม (ไม่มี drag & drop) → แสดงตัวอย่างก่อน แล้วกด **Save photo** ถึงอัปโหลด (Cancel ทิ้งได้) | ปุ่ม (กล้อง/คลังรูป) `accept="image/*"` → ตัวอย่าง + Save photo เหมือนกัน |
| History / ตาราง admin | ตาราง | การ์ด |

- สลับ Dialog ↔ Drawer ด้วย hook `useMediaQuery('(min-width: 768px)')`
- Touch target ≥ 44×44px · input font ≥ 16px (กัน iOS zoom) · `type="tel"` / `type="email"` + `autoComplete`
- รูป `aspect-square object-cover` · ข้อความยาว `break-words` · ไม่มี horizontal scroll · `env(safe-area-inset-bottom)` สำหรับ Drawer

### 9.9 Components & State
| Component | หน้าที่ |
|---|---|
| `AppShell` / `AdminShell` | navbar, sidebar/sheet, guard |
| `ProfileHeader` | รูป + อัปโหลด (preview, ตรวจขนาด/ชนิดฝั่ง client), ชื่อ read-only + 🔒 tooltip |
| `ProfileChecklist` | banner ความครบของโปรไฟล์ (first-time) |
| `ContactField` | edit 1 field (inline / Drawer ตามขนาดจอ), Zod schema, ส่ง `If-Match` |
| `SkillsPicker` | multi-select chips, เทียบกับค่าเดิมเพื่อเปิดปุ่ม Save |
| `RecentChangesCard` | 3 รายการล่าสุด + "View all" + empty state |
| `HistoryTimeline` | group ตามวัน, filter, diff skills, ผู้แก้ |
| `EmptyState` | ไอคอน + หัวข้อ + คำอธิบาย (ใช้ซ้ำทุกหน้า) |

- Server state ด้วย TanStack Query: `useMe`, `useProfile`, `useSkills`, `useHistory`, admin hooks · mutation → `invalidateQueries(['profile','history'])`
- 409 Conflict → toast "This profile was updated elsewhere. Reload to see the latest." + refetch
- Loading skeleton, error state, toast, a11y (label, focus ring, keyboard, aria)

### 9.10 Seed Accounts (สำหรับ demo)
| บัญชี | สถานะ | ใช้ demo |
|---|---|---|
| `EMP-00001` | โปรไฟล์ครบ + history หลายวัน (บางรายการเก่ากว่า 7 วัน) | หน้าปกติ, history ถูกตัดตาม `days` |
| `EMP-00002` | โปรไฟล์ว่าง | empty state + checklist |
| `ADMIN-001` | role admin | หน้า Admin (login แล้วไป `/admin`, ไม่มีเมนู Profile) |

รหัสผ่าน demo อยู่ใน seed file + README (ไม่แสดงบนหน้าจอ) · ข้อมูล seed เป็นข้อมูลสมมติทั้งหมด (ดู §14)

---

## 10. Non-Functional Requirements

| ด้าน | สิ่งที่ทำ |
|---|---|
| Security | JWT httpOnly cookie, bcrypt, rate limit, `MaxBytesReader`, magic-byte check, ไฟล์ตั้งชื่อด้วย UUID (กัน path traversal), CORS allowlist, secure headers, SQL ผ่าน sqlc (parameterized ทั้งหมด), DB role แยก owner/app |
| Data integrity | update + log ใน transaction เดียว (service + `TxManager`), `SELECT ... FOR UPDATE` + optimistic locking (`version`), CHECK constraint, log append-only ระดับสิทธิ์ DB |
| Observability | slog JSON, request-id ใน log + error response, `/healthz` (ping DB) |
| Reliability | graceful shutdown, `pgxpool` settings, timeout ต่อ request (context) |
| Testing | Unit (service, mock repo, table-driven) · Integration (testcontainers Postgres + migrations — transaction, lock, rollback จริง) · FE (Vitest + RTL) · E2E (Playwright: login → แก้เบอร์ → เห็นใน Recent changes / history) รันหลาย viewport (iPhone 13, iPad, 1280px) |
| CI | GitHub Actions: golangci-lint, `sqlc diff` (โค้ด generate ตรงกับ SQL), go test, eslint, vitest, build docker images, gitleaks |
| DevEx | `docker compose up` ครั้งเดียวรันได้ครบ, Makefile, `.env.example`, seed users, OpenAPI spec |

### 10.1 Test Cases สำคัญ
- แก้ contact → version +1 และมี log 1 แถว
- แก้เป็นค่าเดิม → version +1 แต่ **ไม่มี** log
- version ไม่ตรง → 409 และข้อมูล/ log ไม่เปลี่ยน
- 2 request พร้อมกันด้วย version เดียวกัน → สำเร็จ 1, ได้ 409 อีก 1 (ทดสอบ `FOR UPDATE`)
- insert log ล้มเหลว → update ถูก rollback ด้วย (ไม่มีข้อมูลเปลี่ยนโดยไม่มี log)
- field ไม่รู้จัก → 400
- skills มี id ที่ inactive ซึ่งพนักงานยังไม่มี → 400, ไม่มีอะไรเปลี่ยน · ถ้ามีอยู่แล้ว → เก็บไว้ได้
- skills ชุดเดิมแต่สลับลำดับ → version +1 แต่ไม่มี log · insert log ล้มเหลว → skills ถูก rollback
- history: log อายุ 8 วันไม่แสดงเมื่อ `history_days=7` แต่แสดงเมื่อ `=10` และ `hasOlder` ถูกต้อง
- history: ไม่เคยแก้ → `items: []`, `hasOlder: false` · แก้นานแล้ว → `items: []`, `hasOlder: true` · เห็นเฉพาะของตัวเอง
- history: `app_settings` ไม่มีค่าหรือค่าเสีย → ใช้ env `HISTORY_DAYS` (หน้าไม่พัง) · `?field=` ที่ไม่รู้จัก / `page` ไม่ใช่ตัวเลข / `pageSize` > 100 → 400
- employee เรียก `/admin/*` → 403 · เรียก `AdminService` ด้วย actor ที่ไม่ใช่ admin → `ErrForbidden`
- admin แก้ชื่อ → log `first_name`/`last_name` ระบุ `changed_by` = admin
- ปิด skill → ไม่อยู่ใน `/skills` แต่พนักงานที่มีอยู่แล้วยังเห็น · ชื่อ skill ซ้ำ (ไม่สนตัวพิมพ์) → 409
- admin ตั้ง `history_days = 30` → `/profile/history` คืน `days: 30` ทันที (ไม่ restart) · ค่า 0 หรือ 400 → 400
- บัญชี `is_active = FALSE` login ไม่ได้
- role `app` สั่ง `UPDATE`/`DELETE` บน `profile_change_logs` → permission denied
- E2E responsive: แก้ contact ผ่าน Drawer บนมือถือ, ไม่มี horizontal scroll ทุก breakpoint
- พยายามแก้ชื่อ (ไม่มี route) → 404/405
- อัปโหลดไฟล์ .txt เปลี่ยนนามสกุลเป็น .jpg → 415
- DB error ระหว่าง update photo → ไฟล์ใหม่ถูกลบ

---

## 11. docker-compose (โครง)
```yaml
services:
  postgres:
    image: postgres:17-alpine
    environment: { POSTGRES_DB: portal, POSTGRES_USER: owner, POSTGRES_PASSWORD: owner }
    healthcheck: { test: ["CMD-SHELL", "pg_isready -U owner -d portal"], interval: 5s }
    volumes:
      - pg-data:/var/lib/postgresql/data
      - ./backend/db-init:/docker-entrypoint-initdb.d:ro   # สร้าง role app
  migrate:
    image: migrate/migrate:v4.19.1
    depends_on: { postgres: { condition: service_healthy } }
    volumes: [./backend/pkg/database/migrations:/migrations:ro]
    command: ["-path=/migrations", "-database=postgres://owner:owner@postgres:5432/portal?sslmode=disable", "up"]
  api:
    build: ./backend
    env_file: .env
    environment: { DATABASE_URL: "postgres://app:app@postgres:5432/portal?sslmode=disable" }
    depends_on: { migrate: { condition: service_completed_successfully } }
    volumes: [uploads:/data/uploads]
    ports: ["8080:8080"]
  # web: (ขั้น frontend)
volumes: { pg-data: {}, uploads: {} }
```
`make up` สร้าง `.env` (สุ่ม `JWT_SECRET`) ถ้ายังไม่มี แล้วรัน `docker compose up -d --build`
รหัสผ่านใน compose เป็นค่าสำหรับ local dev เท่านั้น (ดู §14)

---

## 12. Design Decisions & Trade-offs (สำหรับ walkthrough)

| การตัดสินใจ | เหตุผล | สิ่งที่แลก |
|---|---|---|
| PostgreSQL + logic ใน Go service (ไม่ใช้ Stored Procedure) | ตรงทิศทางทีม (ย้ายจาก MySQL SP), test/debug ใน Go ง่าย, logic อยู่ที่เดียวกับโค้ด, ไม่ผูกกับ DB | ต้องจัดการ transaction เองใน service → `TxManager` + integration test |
| Gin (ไม่ใช่ Fiber) | อยู่บน `net/http` มาตรฐาน: ใช้ `httptest` และ middleware ของ ecosystem ได้, ส่ง `context.Context` ลง pgx ตรง ๆ, ไม่มี pitfall เรื่อง buffer reuse ของ fasthttp | benchmark ช้ากว่า Fiber — ไม่มีผลเพราะคอขวดอยู่ที่ DB และการ resize รูป |
| sqlc แทน ORM (GORM) | SQL ชัดเจน อ่าน/อธิบายได้ทุกบรรทัด, type-safe, query ผิดเจอตอน generate (CI: `sqlc diff`), ไม่มีพฤติกรรมแฝง (N+1, zero-value), ใช้ feature ของ Postgres ได้เต็มที่ (`FOR UPDATE`, `RETURNING`, JSONB) | dynamic query ทำยากกว่า → ใช้ `sqlc.narg` + CASE · ถ้าต้อง dynamic จริงใช้ pgx `CollectRows` ใน repository เดียวกัน · sqlc อยู่แค่ใน `repository.go` จึงเปลี่ยนเป็น GORM ได้โดยไม่แตะ use case |
| Clean Architecture แบบ Package by Feature (1 module = 1 package) | แก้/เพิ่ม/ลบฟีเจอร์อยู่ใน folder เดียว, ชื่อสั้นไม่ stutter (`profile.NewUsecase`), กฎธุรกิจอยู่ใน `entity` ทดสอบได้โดยไม่มี DB | คอมไพเลอร์ไม่บังคับ layer ภายใน module → คุมด้วย interface · module ใช้ตารางร่วมกัน (ไม่ใช่ modular monolith เต็มรูป) เพราะ update + log ต้องอยู่ transaction เดียว |
| Transaction + `FOR UPDATE` + `version` | กันแก้ทับกัน, แยก 404/409 ชัด, log ไม่หลุดจาก update | lock แถวสั้น ๆ ระหว่าง transaction (ไม่มีผลที่ scale นี้) |
| History เขียนใน service (ไม่ใช้ DB trigger) | diff ที่อ่านง่าย (skills เป็นชื่อ), รู้ว่าใครแก้จาก context, test ใน Go ได้ | ถ้ามีคนแก้ DB ตรง ๆ จะไม่มี log → ชดเชยด้วยสิทธิ์ DB role `app` (trigger เป็นทางเลือกในอนาคต) |
| History append-only ด้วยสิทธิ์ DB | audit เชื่อถือได้ แม้แอปมีบั๊ก | purge ต้องใช้ role แยก |
| JSONB สำหรับ `old_value`/`new_value` | ค่าเดี่ยวและ array (skills) อยู่ column เดียว, FE ได้ type ตรง | query ข้างในค่ายากกว่า TEXT เล็กน้อย (ไม่จำเป็นต้องทำ) |
| Validation 2 ชั้น (Go + CHECK constraint) | Go ให้ error ที่ผู้ใช้อ่านเข้าใจ, DB กันข้อมูลเสียจากทุกช่องทาง | กฎซ้ำ 2 ที่ |
| Contact ใช้ endpoint เดียว `PUT /profile/contact/:field` + map กฎต่อ field (table-driven) | 1 handler/service, เพิ่ม field ใหม่ง่าย, ยัง 1 request = 1 field = 1 log ตรงโจทย์ | ไม่ใช้ PATCH แบบ partial body (เลี่ยงความกำกวม "ไม่ส่ง" vs "ส่งค่าว่าง") |
| Log อยู่ใน PostgreSQL เดียวกัน (ไม่แยก Mongo) | atomic กับการ update, ไม่มีปัญหา dual write, ปริมาณข้อมูลน้อย | ถ้าโตมาก → Outbox pattern ส่งต่อ Mongo/Elasticsearch โดยเปลี่ยนแค่ history repository |
| "7 วัน" เป็นช่วงแสดง ไม่ใช่ TTL | เปลี่ยน config แล้วข้อมูลเก่ายังอยู่ | ข้อมูลโตต่อเนื่อง → purge job แยก (optional) |
| `history_days` อยู่ใน DB (`app_settings`) + env เป็น fallback | admin ปรับได้ทันทีโดยไม่ deploy | อ่านค่าทุก request (เล็กมาก, cache ได้ภายหลัง) |
| RBAC ตรวจทั้ง middleware และ service | defense in depth — ถ้าลืมใส่ middleware ที่ route ใหม่ service ยังกันอยู่ | ตรวจซ้ำ 2 ที่ |
| Soft disable skill แทนการลบ | ข้อมูลเก่าของพนักงานไม่พัง, history ยังอ่านได้ | ตารางมี row ที่ไม่ใช้สะสม |
| เก็บรูปบน local volume | ง่าย, demo ได้ทันที | scale หลายเครื่องไม่ได้ → `FileStorage` interface พร้อมเปลี่ยนเป็น S3/MinIO |
| ไม่มีหน้า onboarding แยก — ใช้ empty state + checklist | เข้าใช้งานได้ทันที, ไม่บังคับ wizard, โค้ดน้อยกว่า | ผู้ใช้อาจไม่กรอกครบ (checklist ช่วยกระตุ้น) |
| History: การ์ดสรุปในหน้า Profile + หน้าเต็มผ่าน "View all" (ไม่มีเมนูบน navbar) | navigation เรียบ, history เป็นส่วนหนึ่งของโปรไฟล์, เห็นผลทันทีหลังแก้ | ต้องสื่อตอน demo ว่าหน้าเต็มอยู่ที่ "View all" |
| Vite SPA (ไม่ใช่ Next.js) | เร็ว, ง่าย, เข้ากับ Go API | ไม่มี SSR/SEO — ไม่จำเป็นสำหรับ portal ภายใน |

---

## 13. Implementation Plan

| ขั้น | งาน | Done เมื่อ |
|---|---|---|
| 1 | Repo (public), docker-compose, Makefile, `.env.example`, `.gitignore`, LICENSE | `make up` แล้ว Postgres healthy |
| 2 | Migration: tables + roles/grants + seed users/skills/settings | `make migrate seed` ผ่าน |
| 3 | SQL queries + `sqlc generate` + `TxManager` | `go build` ผ่าน, `sqlc diff` สะอาด |
| 4 | Go skeleton: config, router, middleware, error envelope, `/healthz` | `curl /healthz` ok |
| 5 | **Vertical slice:** `GET /profile` + `PUT /profile/contact/:field` (service + transaction + log) ครบทุกชั้น + หน้า Profile | แก้เบอร์จาก UI ได้จริง และเห็น log |
| 6 | Auth (login/logout/me/middleware) | ต้อง login ก่อนเข้า `/profile/*` |
| 7 | Endpoint ที่เหลือ: skills, photo, history | ครบ F1–F6 |
| 8 | Frontend: SkillsPicker, Photo upload, Recent changes, History page, empty state, responsive | ครบทุกหน้า ทั้ง desktop และมือถือ |
| 9 | Tests (unit + integration + E2E หลาย viewport) | CI เขียว |
| 10 | README (screenshots, วิธีรัน, demo accounts, design decisions), OpenAPI, ซ้อม demo / walkthrough | **พร้อมส่ง (core เสร็จ)** |
| 11 | Admin Tier 1: role + RBAC, Skills, Settings | ปรับ skill / history days จากหน้าจอได้ |
| 12 | Admin Tier 2: Employees list/detail, แก้ชื่อ, Dashboard | admin แก้ชื่อแล้วเห็นใน history |
| 13 | Admin Tier 3: สร้างบัญชี, reset รหัส, must-change-password | ถ้ามีเวลา |

### Nice-to-have (ถ้ามีเวลา)
- Crop รูปก่อนอัปโหลด, dark mode, i18n TH/EN
- History: ปุ่ม Revert, แจ้งเตือนเมื่ออีเมล/เบอร์ถูกเปลี่ยน, Export CSV (admin)
- Scheduled purge history
- Outbox → Mongo (เพื่อโชว์แนวคิด)

---

## 14. Public Repository Checklist

Repo เป็น public — ใครก็อ่านได้ทั้งโค้ดและ commit history

| เรื่อง | สิ่งที่ทำ |
|---|---|
| Secrets | ไม่ commit `.env` (อยู่ใน `.gitignore`), มีแค่ `.env.example` · `JWT_SECRET` ไม่มีค่า default ในโค้ด (ไม่ตั้ง = แอปไม่ start) · รหัสผ่านใน compose/seed เป็นค่า local dev ที่ระบุชัดว่าใช้ demo เท่านั้น |
| Secret scanning | เปิด GitHub secret scanning + push protection · `gitleaks` ใน CI · ถ้าเผลอ commit secret → **เปลี่ยน secret ทันที** (ลบออกจาก history อย่างเดียวไม่พอ) |
| ข้อมูลส่วนบุคคล | seed ใช้ชื่อ/เบอร์/อีเมลสมมติ (`@example.com`) · ไม่ commit รูปคนจริง · ไม่ commit ไฟล์ใน `uploads/` |
| เอกสารของบริษัท | ไม่ commit ไฟล์ PDF โจทย์ — สรุป requirement ด้วยคำของตัวเองใน README |
| Dependencies | เปิด Dependabot (Go modules, npm, GitHub Actions) |
| License | ใส่ `LICENSE` (เช่น MIT) |
| Commit history | commit ย่อยและข้อความสื่อความ (กรรมการอาจอ่าน history) · ไม่มี commit "fix", "wip" รัว ๆ บน `main` |
| README | ภาพหน้าจอ desktop + mobile, วิธีรันใน 1 คำสั่ง, demo accounts, สรุป design decisions + ลิงก์ `docs/DESIGN.md` |
