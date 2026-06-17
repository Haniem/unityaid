CREATE TABLE IF NOT EXISTS certificates (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	organization_id UUID REFERENCES organizations(id) ON DELETE SET NULL,
	type TEXT NOT NULL DEFAULT 'participation',
	title TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	total_hours NUMERIC(10, 2) NOT NULL DEFAULT 0,
	verify_code TEXT NOT NULL UNIQUE,
	file_path TEXT NOT NULL DEFAULT '',
	issued_by UUID REFERENCES users(id) ON DELETE SET NULL,
	issued_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_certificates_user_id ON certificates(user_id);
CREATE INDEX IF NOT EXISTS idx_certificates_verify_code ON certificates(verify_code);
