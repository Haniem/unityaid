-- Ядро платформы (основная диаграмма для диплома)
-- Таблиц: 10

CREATE TABLE organizations (
  id uuid NOT NULL,
  name text NOT NULL,
  PRIMARY KEY (id)
);

CREATE TABLE users (
  id uuid NOT NULL,
  email text NOT NULL,
  first_name text NOT NULL,
  last_name text NOT NULL,
  PRIMARY KEY (id)
);

CREATE TABLE volunteer_profiles (
  id uuid NOT NULL,
  user_id uuid NOT NULL REFERENCES users(id),
  city text,
  points integer NOT NULL,
  level integer NOT NULL,
  status text NOT NULL,
  PRIMARY KEY (id)
);

CREATE TABLE events (
  id uuid NOT NULL,
  organization_id uuid NOT NULL REFERENCES organizations(id),
  title text NOT NULL,
  format text NOT NULL,
  status text NOT NULL,
  starts_at timestamptz NOT NULL,
  ends_at timestamptz NOT NULL,
  created_by uuid REFERENCES users(id),
  PRIMARY KEY (id)
);

CREATE TABLE organization_members (
  id uuid NOT NULL,
  organization_id uuid NOT NULL REFERENCES organizations(id),
  user_id uuid NOT NULL REFERENCES users(id),
  role text NOT NULL,
  status text NOT NULL,
  PRIMARY KEY (id)
);

CREATE TABLE tasks (
  id uuid NOT NULL,
  organization_id uuid NOT NULL REFERENCES organizations(id),
  event_id uuid REFERENCES events(id),
  title text NOT NULL,
  status text NOT NULL,
  priority text NOT NULL,
  due_at timestamptz,
  created_by uuid REFERENCES users(id),
  completion_confirmed_by uuid REFERENCES users(id),
  PRIMARY KEY (id)
);

CREATE TABLE event_attendance (
  id uuid NOT NULL,
  event_id uuid NOT NULL REFERENCES events(id),
  user_id uuid NOT NULL REFERENCES users(id),
  check_in_at timestamptz,
  check_out_at timestamptz,
  hours numeric NOT NULL,
  PRIMARY KEY (id)
);

CREATE TABLE event_applications (
  id uuid NOT NULL,
  event_id uuid NOT NULL REFERENCES events(id),
  user_id uuid NOT NULL REFERENCES users(id),
  status text NOT NULL,
  PRIMARY KEY (id)
);

CREATE TABLE task_assignments (
  id uuid NOT NULL,
  task_id uuid NOT NULL REFERENCES tasks(id),
  user_id uuid NOT NULL REFERENCES users(id),
  assigned_at timestamptz NOT NULL,
  role text NOT NULL,
  PRIMARY KEY (id)
);

CREATE TABLE time_entries (
  id uuid NOT NULL,
  organization_id uuid NOT NULL REFERENCES organizations(id),
  user_id uuid NOT NULL REFERENCES users(id),
  event_id uuid REFERENCES events(id),
  task_id uuid REFERENCES tasks(id),
  hours numeric NOT NULL,
  status text NOT NULL,
  reviewed_by uuid REFERENCES users(id),
  PRIMARY KEY (id)
);
