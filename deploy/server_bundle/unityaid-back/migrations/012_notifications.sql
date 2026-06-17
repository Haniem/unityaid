CREATE TABLE IF NOT EXISTS notifications (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	type TEXT NOT NULL,
	title TEXT NOT NULL,
	body TEXT NOT NULL DEFAULT '',
	link TEXT NOT NULL DEFAULT '',
	entity_type TEXT NOT NULL DEFAULT '',
	entity_id TEXT NOT NULL DEFAULT '',
	is_read BOOLEAN NOT NULL DEFAULT false,
	read_at TIMESTAMPTZ,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_notifications_dedupe
	ON notifications(user_id, type, entity_type, entity_id);

CREATE INDEX IF NOT EXISTS idx_notifications_user_created
	ON notifications(user_id, is_read, created_at DESC);

CREATE OR REPLACE FUNCTION create_achievement_notification()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
	achievement_name TEXT;
BEGIN
	SELECT name INTO achievement_name FROM achievements WHERE id = NEW.achievement_id;

	INSERT INTO notifications (user_id, type, title, body, link, entity_type, entity_id)
	VALUES (
		NEW.user_id,
		'achievement_earned',
		'Достижение получено',
		COALESCE(achievement_name, ''),
		'/achievements',
		'achievement',
		NEW.achievement_id::text
	)
	ON CONFLICT (user_id, type, entity_type, entity_id)
	DO UPDATE SET
		title = EXCLUDED.title,
		body = EXCLUDED.body,
		link = EXCLUDED.link,
		is_read = false,
		read_at = NULL,
		created_at = now();

	RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_achievement_notification ON volunteer_achievements;
CREATE TRIGGER trg_achievement_notification
AFTER INSERT ON volunteer_achievements
FOR EACH ROW
EXECUTE FUNCTION create_achievement_notification();
