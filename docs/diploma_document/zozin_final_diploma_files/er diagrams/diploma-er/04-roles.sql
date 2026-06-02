-- Роли и членство (приложение)
-- Таблиц: 5

CREATE TABLE organizations (
  id uuid NOT NULL,
  name text NOT NULL,
  PRIMARY KEY (id)
);

CREATE TABLE system_roles (
  id uuid NOT NULL,
  code text NOT NULL,
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

CREATE TABLE user_system_roles (
  user_id uuid NOT NULL REFERENCES users(id),
  role_id uuid NOT NULL REFERENCES system_roles(id),
  assigned_at timestamptz NOT NULL,
  PRIMARY KEY (user_id, role_id)
);

CREATE TABLE organization_members (
  id uuid NOT NULL,
  organization_id uuid NOT NULL REFERENCES organizations(id),
  user_id uuid NOT NULL REFERENCES users(id),
  role text NOT NULL,
  status text NOT NULL,
  PRIMARY KEY (id)
);
