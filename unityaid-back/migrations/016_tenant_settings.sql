CREATE TABLE IF NOT EXISTS tenant_settings (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	display_name TEXT NOT NULL DEFAULT 'UnityAid',
	description TEXT NOT NULL DEFAULT '',
	logo_url TEXT,
	primary_color TEXT NOT NULL DEFAULT '#2f9f72',
	accent_color TEXT NOT NULL DEFAULT '#22684e',
	timezone TEXT NOT NULL DEFAULT 'Asia/Yekaterinburg',
	locale TEXT NOT NULL DEFAULT 'ru',
	contact_email TEXT,
	contact_phone TEXT,
	default_organization_id UUID REFERENCES organizations(id) ON DELETE SET NULL,
	default_organization_name TEXT NOT NULL DEFAULT '',
	default_organization_slug TEXT NOT NULL DEFAULT '',
	onboarding_completed BOOLEAN NOT NULL DEFAULT false,
	pending_invites JSONB NOT NULL DEFAULT '[]'::jsonb,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	singleton_key BOOLEAN NOT NULL DEFAULT true,
	CONSTRAINT tenant_settings_singleton CHECK (singleton_key),
	CONSTRAINT tenant_settings_singleton_unique UNIQUE (singleton_key)
);

INSERT INTO tenant_settings (singleton_key)
VALUES (true)
ON CONFLICT (singleton_key) DO NOTHING;
