CREATE TABLE IF NOT EXISTS news (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	organization_id UUID REFERENCES organizations(id) ON DELETE SET NULL,
	title TEXT NOT NULL,
	slug TEXT NOT NULL UNIQUE,
	summary TEXT NOT NULL DEFAULT '',
	content_html TEXT NOT NULL DEFAULT '',
	cover_image_url TEXT,
	status TEXT NOT NULL DEFAULT 'published',
	author_id UUID REFERENCES users(id) ON DELETE SET NULL,
	published_at TIMESTAMPTZ,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_news_organization_id ON news(organization_id);
CREATE INDEX IF NOT EXISTS idx_news_created_at ON news(created_at DESC);
