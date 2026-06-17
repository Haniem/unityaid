ALTER TABLE event_applications
	ADD COLUMN IF NOT EXISTS rejection_reason TEXT NOT NULL DEFAULT '';

CREATE TABLE IF NOT EXISTS event_templates (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
	name TEXT NOT NULL,
	title TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	format event_format NOT NULL DEFAULT 'offline',
	location TEXT,
	max_participants INTEGER,
	default_duration_minutes INTEGER NOT NULL DEFAULT 120,
	created_by UUID REFERENCES users(id) ON DELETE SET NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_event_templates_organization_id ON event_templates(organization_id);
