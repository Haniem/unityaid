CREATE TABLE IF NOT EXISTS time_entries (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
	user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	event_id UUID REFERENCES events(id) ON DELETE SET NULL,
	task_id UUID REFERENCES tasks(id) ON DELETE SET NULL,
	hours NUMERIC(8, 2) NOT NULL CHECK (hours > 0),
	description TEXT NOT NULL DEFAULT '',
	status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
	reviewed_by UUID REFERENCES users(id) ON DELETE SET NULL,
	reviewed_at TIMESTAMPTZ,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	CHECK (event_id IS NOT NULL OR task_id IS NOT NULL)
);

CREATE INDEX IF NOT EXISTS idx_time_entries_user_id ON time_entries(user_id);
CREATE INDEX IF NOT EXISTS idx_time_entries_organization_id ON time_entries(organization_id);
CREATE INDEX IF NOT EXISTS idx_time_entries_status ON time_entries(status);
