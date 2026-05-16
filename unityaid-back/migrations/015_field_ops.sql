CREATE TABLE IF NOT EXISTS field_qr_tokens (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
	token TEXT NOT NULL UNIQUE,
	mode TEXT NOT NULL DEFAULT 'checkin' CHECK (mode IN ('checkin', 'checkout', 'attendance')),
	expires_at TIMESTAMPTZ,
	created_by UUID REFERENCES users(id) ON DELETE SET NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS field_checkins (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
	user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	qr_token_id UUID REFERENCES field_qr_tokens(id) ON DELETE SET NULL,
	checkin_at TIMESTAMPTZ,
	checkout_at TIMESTAMPTZ,
	status TEXT NOT NULL DEFAULT 'checked_in' CHECK (status IN ('checked_in', 'checked_out')),
	source TEXT NOT NULL DEFAULT 'qr',
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	UNIQUE (event_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_field_qr_tokens_event_id ON field_qr_tokens(event_id);
CREATE INDEX IF NOT EXISTS idx_field_qr_tokens_token ON field_qr_tokens(token);
CREATE INDEX IF NOT EXISTS idx_field_checkins_event_id ON field_checkins(event_id);
CREATE INDEX IF NOT EXISTS idx_field_checkins_user_id ON field_checkins(user_id);
