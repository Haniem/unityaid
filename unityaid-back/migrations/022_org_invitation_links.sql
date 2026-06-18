ALTER TABLE user_invitations
	DROP CONSTRAINT IF EXISTS user_invitations_email_check;

ALTER TABLE user_invitations
	ALTER COLUMN email DROP NOT NULL,
	ADD COLUMN IF NOT EXISTS organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_user_invitations_organization_id ON user_invitations(organization_id);
