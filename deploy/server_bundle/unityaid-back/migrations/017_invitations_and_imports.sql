CREATE TABLE IF NOT EXISTS user_invitations (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	email TEXT NOT NULL,
	role TEXT NOT NULL DEFAULT 'volunteer',
	token TEXT NOT NULL UNIQUE,
	status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'revoked', 'expired')),
	invited_by UUID REFERENCES users(id) ON DELETE SET NULL,
	accepted_by UUID REFERENCES users(id) ON DELETE SET NULL,
	expires_at TIMESTAMPTZ NOT NULL DEFAULT (now() + interval '14 days'),
	accepted_at TIMESTAMPTZ,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_user_invitations_email ON user_invitations(lower(email));
CREATE INDEX IF NOT EXISTS idx_user_invitations_token ON user_invitations(token);
