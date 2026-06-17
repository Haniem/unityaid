ALTER TABLE organizations
	ADD COLUMN IF NOT EXISTS logo_url TEXT,
	ADD COLUMN IF NOT EXISTS website_url TEXT,
	ADD COLUMN IF NOT EXISTS phone TEXT,
	ADD COLUMN IF NOT EXISTS address TEXT,
	ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_organizations_deleted_at ON organizations(deleted_at);
CREATE INDEX IF NOT EXISTS idx_organizations_name_search ON organizations USING gin (to_tsvector('simple', name || ' ' || slug || ' ' || description));
