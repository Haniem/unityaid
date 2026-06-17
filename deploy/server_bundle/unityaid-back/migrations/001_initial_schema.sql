CREATE EXTENSION IF NOT EXISTS pgcrypto;

DO $$
BEGIN
	IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'app_role') THEN
		CREATE TYPE app_role AS ENUM ('super_admin', 'org_admin', 'coordinator', 'volunteer');
	END IF;

	IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'member_status') THEN
		CREATE TYPE member_status AS ENUM ('active', 'inactive', 'blocked');
	END IF;

	IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'event_format') THEN
		CREATE TYPE event_format AS ENUM ('online', 'offline', 'hybrid');
	END IF;

	IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'event_status') THEN
		CREATE TYPE event_status AS ENUM ('draft', 'published', 'completed', 'cancelled');
	END IF;

	IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'task_status') THEN
		CREATE TYPE task_status AS ENUM ('created', 'assigned', 'in_progress', 'review', 'completed', 'cancelled');
	END IF;
END $$;

CREATE TABLE IF NOT EXISTS organizations (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	name TEXT NOT NULL,
	slug TEXT NOT NULL UNIQUE,
	description TEXT NOT NULL DEFAULT '',
	contact_email TEXT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS users (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	email TEXT NOT NULL UNIQUE,
	password_hash TEXT NOT NULL,
	first_name TEXT NOT NULL,
	last_name TEXT NOT NULL,
	patronymic TEXT,
	avatar_url TEXT,
	locale TEXT NOT NULL DEFAULT 'ru',
	is_email_verified BOOLEAN NOT NULL DEFAULT true,
	is_active BOOLEAN NOT NULL DEFAULT true,
	last_login_at TIMESTAMPTZ,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS organization_members (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
	user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	role app_role NOT NULL,
	status member_status NOT NULL DEFAULT 'active',
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	UNIQUE (organization_id, user_id)
);

CREATE TABLE IF NOT EXISTS volunteer_profiles (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
	city TEXT,
	phone TEXT,
	bio TEXT NOT NULL DEFAULT '',
	total_hours NUMERIC(8, 2) NOT NULL DEFAULT 0,
	points INTEGER NOT NULL DEFAULT 0,
	level INTEGER NOT NULL DEFAULT 1,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS skills (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	name TEXT NOT NULL UNIQUE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS volunteer_skills (
	volunteer_profile_id UUID NOT NULL REFERENCES volunteer_profiles(id) ON DELETE CASCADE,
	skill_id UUID NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
	PRIMARY KEY (volunteer_profile_id, skill_id)
);

CREATE TABLE IF NOT EXISTS events (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
	title TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	format event_format NOT NULL,
	status event_status NOT NULL DEFAULT 'draft',
	starts_at TIMESTAMPTZ NOT NULL,
	ends_at TIMESTAMPTZ NOT NULL,
	location TEXT,
	max_participants INTEGER,
	created_by UUID REFERENCES users(id) ON DELETE SET NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS tasks (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
	event_id UUID REFERENCES events(id) ON DELETE SET NULL,
	title TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	status task_status NOT NULL DEFAULT 'created',
	priority TEXT NOT NULL DEFAULT 'medium',
	due_at TIMESTAMPTZ,
	created_by UUID REFERENCES users(id) ON DELETE SET NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS task_assignments (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
	user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	assigned_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	UNIQUE (task_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_organization_members_user_id ON organization_members(user_id);
CREATE INDEX IF NOT EXISTS idx_events_organization_id ON events(organization_id);
CREATE INDEX IF NOT EXISTS idx_tasks_organization_id ON tasks(organization_id);
