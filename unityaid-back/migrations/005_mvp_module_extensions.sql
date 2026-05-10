ALTER TABLE events
	ADD COLUMN IF NOT EXISTS checkin_code TEXT NOT NULL DEFAULT encode(gen_random_bytes(12), 'hex');

DO $$
BEGIN
	IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'event_application_status') THEN
		CREATE TYPE event_application_status AS ENUM ('pending', 'approved', 'waitlisted', 'rejected', 'cancelled');
	END IF;
END $$;

CREATE TABLE IF NOT EXISTS event_applications (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
	user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	status event_application_status NOT NULL DEFAULT 'pending',
	message TEXT NOT NULL DEFAULT '',
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	UNIQUE (event_id, user_id)
);

CREATE TABLE IF NOT EXISTS event_attendance (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
	user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	check_in_at TIMESTAMPTZ,
	check_out_at TIMESTAMPTZ,
	hours NUMERIC(8, 2) NOT NULL DEFAULT 0,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	UNIQUE (event_id, user_id)
);

CREATE TABLE IF NOT EXISTS event_shifts (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
	title TEXT NOT NULL,
	starts_at TIMESTAMPTZ NOT NULL,
	ends_at TIMESTAMPTZ NOT NULL,
	capacity INTEGER,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS event_feedback (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
	user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	rating INTEGER NOT NULL CHECK (rating BETWEEN 1 AND 5),
	comment TEXT NOT NULL DEFAULT '',
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	UNIQUE (event_id, user_id)
);

ALTER TABLE task_assignments
	ADD COLUMN IF NOT EXISTS role TEXT NOT NULL DEFAULT 'assignee';

ALTER TABLE tasks
	ADD COLUMN IF NOT EXISTS completion_confirmed_by UUID REFERENCES users(id) ON DELETE SET NULL,
	ADD COLUMN IF NOT EXISTS completion_confirmed_at TIMESTAMPTZ;

CREATE TABLE IF NOT EXISTS task_comments (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
	user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	content TEXT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS task_attachments (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
	user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	file_name TEXT NOT NULL,
	file_url TEXT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS task_status_history (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
	from_status TEXT,
	to_status TEXT NOT NULL,
	changed_by UUID REFERENCES users(id) ON DELETE SET NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS task_time_entries (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
	user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	hours NUMERIC(8, 2) NOT NULL,
	note TEXT NOT NULL DEFAULT '',
	status TEXT NOT NULL DEFAULT 'pending',
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS news_categories (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	name TEXT NOT NULL UNIQUE,
	slug TEXT NOT NULL UNIQUE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE news
	ADD COLUMN IF NOT EXISTS category_id UUID REFERENCES news_categories(id) ON DELETE SET NULL,
	ADD COLUMN IF NOT EXISTS scheduled_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_event_applications_event_id ON event_applications(event_id);
CREATE INDEX IF NOT EXISTS idx_event_attendance_event_id ON event_attendance(event_id);
CREATE INDEX IF NOT EXISTS idx_task_comments_task_id ON task_comments(task_id);
CREATE INDEX IF NOT EXISTS idx_task_status_history_task_id ON task_status_history(task_id);
CREATE INDEX IF NOT EXISTS idx_news_status ON news(status);
CREATE INDEX IF NOT EXISTS idx_news_scheduled_at ON news(scheduled_at);
