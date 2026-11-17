CREATE TABLE users (
  id                    BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  employee_code         TEXT        NOT NULL UNIQUE CHECK (length(employee_code) BETWEEN 3 AND 20),
  password_hash         TEXT        NOT NULL,
  role                  TEXT        NOT NULL DEFAULT 'employee' CHECK (role IN ('employee', 'admin')),
  is_active             BOOLEAN     NOT NULL DEFAULT TRUE,
  must_change_password  BOOLEAN     NOT NULL DEFAULT FALSE,
  created_at            TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE app_settings (
  key         TEXT PRIMARY KEY,
  value       TEXT        NOT NULL,
  updated_by  BIGINT REFERENCES users (id),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE employees (
  id             BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  user_id        BIGINT      NOT NULL UNIQUE REFERENCES users (id),
  first_name     TEXT        NOT NULL CHECK (length(first_name) BETWEEN 1 AND 100),
  last_name      TEXT        NOT NULL CHECK (length(last_name) BETWEEN 1 AND 100),
  photo_path     TEXT,
  mobile_phone   TEXT        CHECK (mobile_phone ~ '^\+?[0-9]{9,15}$'),
  contact_email  TEXT        CHECK (length(contact_email) <= 254),
  address        TEXT        CHECK (length(address) BETWEEN 1 AND 500),
  version        INTEGER     NOT NULL DEFAULT 1,
  updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE skills (
  id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  name        TEXT    NOT NULL CHECK (length(name) BETWEEN 1 AND 50),
  is_active   BOOLEAN NOT NULL DEFAULT TRUE,
  sort_order  INTEGER NOT NULL DEFAULT 0
);
CREATE UNIQUE INDEX uq_skills_name ON skills (lower(name));

CREATE TABLE employee_skills (
  employee_id BIGINT NOT NULL REFERENCES employees (id) ON DELETE CASCADE,
  skill_id    BIGINT NOT NULL REFERENCES skills (id),
  PRIMARY KEY (employee_id, skill_id)
);

CREATE TABLE profile_change_logs (
  id           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  employee_id  BIGINT      NOT NULL REFERENCES employees (id) ON DELETE CASCADE,
  field        TEXT        NOT NULL CHECK (field IN ('first_name', 'last_name', 'mobile_phone',
                                                     'contact_email', 'address', 'photo', 'skills')),
  old_value    JSONB,
  new_value    JSONB,
  changed_by   BIGINT      NOT NULL REFERENCES users (id),
  changed_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_log_emp_time ON profile_change_logs (employee_id, changed_at DESC);
