DO $$
BEGIN
	IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'volunteer_status') THEN
		CREATE TYPE volunteer_status AS ENUM ('new', 'active', 'unavailable', 'archived');
	END IF;
END $$;

ALTER TABLE volunteer_profiles
	ADD COLUMN IF NOT EXISTS status volunteer_status NOT NULL DEFAULT 'active',
	ADD COLUMN IF NOT EXISTS interests TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_volunteer_profiles_status ON volunteer_profiles(status);
