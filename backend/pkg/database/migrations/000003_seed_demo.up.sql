-- Demo data for local development and the assignment walkthrough.
-- All names, phone numbers and emails are fictional. Password for every account: demo1234

INSERT INTO app_settings (key, value) VALUES ('history_days', '7');

INSERT INTO skills (name, sort_order) VALUES
  ('Go', 1), ('React', 2), ('Flutter', 3), ('CI/CD', 4),
  ('Docker', 5), ('Kubernetes', 6), ('PostgreSQL', 7), ('TypeScript', 8);

INSERT INTO users (employee_code, password_hash, role) VALUES
  ('EMP-00001', '$2a$10$NQ.nWEZwH8//fqJvQwKneub8Nkn7s1pH91J2sPwMMz3ProcdR9y1S', 'employee'),
  ('EMP-00002', '$2a$10$NQ.nWEZwH8//fqJvQwKneub8Nkn7s1pH91J2sPwMMz3ProcdR9y1S', 'employee'),
  ('ADMIN-001', '$2a$10$NQ.nWEZwH8//fqJvQwKneub8Nkn7s1pH91J2sPwMMz3ProcdR9y1S', 'admin');

-- EMP-00001: complete profile with history inside and outside the 7-day window.
INSERT INTO employees (user_id, first_name, last_name, mobile_phone, contact_email, address, version)
SELECT id, 'Somchai', 'Jaidee', '0812345678', 'somchai@example.com',
       '99 Example Road, Bangkok 10110', 6
FROM users WHERE employee_code = 'EMP-00001';

-- EMP-00002: empty profile for the first-time experience.
INSERT INTO employees (user_id, first_name, last_name)
SELECT id, 'Suda', 'Rakdee' FROM users WHERE employee_code = 'EMP-00002';

INSERT INTO employees (user_id, first_name, last_name)
SELECT id, 'Admin', 'User' FROM users WHERE employee_code = 'ADMIN-001';

INSERT INTO employee_skills (employee_id, skill_id)
SELECT e.id, s.id
FROM employees e
JOIN users u ON u.id = e.user_id
JOIN skills s ON s.name IN ('Go', 'React', 'CI/CD')
WHERE u.employee_code = 'EMP-00001';

INSERT INTO profile_change_logs (employee_id, field, old_value, new_value, changed_by, changed_at)
SELECT e.id, v.field, v.old_value::jsonb, v.new_value::jsonb,
       CASE WHEN v.by_admin THEN a.id ELSE u.id END,
       now() - v.ago
FROM employees e
JOIN users u ON u.id = e.user_id
CROSS JOIN (SELECT id FROM users WHERE employee_code = 'ADMIN-001') a
CROSS JOIN (VALUES
  ('contact_email', NULL,                    '"somchai.j@example.com"', FALSE, interval '20 days'),
  ('last_name',     '"Jaidi"',               '"Jaidee"',                TRUE,  interval '10 days'),
  ('address',       NULL,                    '"99 Example Road, Bangkok 10110"', FALSE, interval '9 days'),
  ('contact_email', '"somchai.j@example.com"', '"somchai@example.com"', FALSE, interval '3 days'),
  ('skills',        '["Go","React","Flutter"]', '["Go","React","CI/CD"]', FALSE, interval '1 day'),
  ('mobile_phone',  '"0811111111"',          '"0812345678"',            FALSE, interval '2 hours')
) AS v(field, old_value, new_value, by_admin, ago)
WHERE u.employee_code = 'EMP-00001';
