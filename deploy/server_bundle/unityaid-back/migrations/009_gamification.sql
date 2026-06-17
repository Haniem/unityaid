CREATE TABLE IF NOT EXISTS achievements (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	code TEXT NOT NULL UNIQUE,
	name TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	icon TEXT NOT NULL DEFAULT 'award',
	points_reward INTEGER NOT NULL DEFAULT 0,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS achievement_rules (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	achievement_id UUID NOT NULL UNIQUE REFERENCES achievements(id) ON DELETE CASCADE,
	code TEXT NOT NULL UNIQUE,
	trigger_type TEXT NOT NULL,
	threshold NUMERIC(10, 2),
	points INTEGER NOT NULL DEFAULT 0,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS volunteer_achievements (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	achievement_id UUID NOT NULL REFERENCES achievements(id) ON DELETE CASCADE,
	earned_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	UNIQUE (user_id, achievement_id)
);

CREATE TABLE IF NOT EXISTS points_transactions (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	achievement_id UUID REFERENCES achievements(id) ON DELETE SET NULL,
	source_type TEXT NOT NULL DEFAULT 'achievement',
	source_id UUID,
	points INTEGER NOT NULL,
	reason TEXT NOT NULL DEFAULT '',
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	UNIQUE (user_id, achievement_id, source_type)
);

INSERT INTO achievements (code, name, description, icon, points_reward)
VALUES
	('registration', 'Первый шаг', 'Регистрация в системе «Пульс».', 'user-plus', 20),
	('profile_completed', 'Профиль заполнен', 'Заполнены основные данные профиля волонтера.', 'id-card', 30),
	('first_participation', 'Первое участие', 'Волонтер впервые отмечен на мероприятии.', 'calendar-check', 50),
	('task_completed', 'Задача выполнена', 'Волонтер участвовал в выполненной задаче.', 'check-circle', 50),
	('hours_10', '10 часов помощи', 'Накоплено 10 подтвержденных волонтерских часов.', 'clock', 100),
	('hours_25', '25 часов помощи', 'Накоплено 25 подтвержденных волонтерских часов.', 'clock', 180),
	('hours_50', '50 часов помощи', 'Накоплено 50 подтвержденных волонтерских часов.', 'trophy', 300),
	('hours_100', '100 часов помощи', 'Накоплено 100 подтвержденных волонтерских часов.', 'medal', 500)
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name,
	description = EXCLUDED.description,
	icon = EXCLUDED.icon,
	points_reward = EXCLUDED.points_reward;

INSERT INTO achievement_rules (achievement_id, code, trigger_type, threshold, points)
SELECT id, code, code, CASE
	WHEN code = 'hours_10' THEN 10
	WHEN code = 'hours_25' THEN 25
	WHEN code = 'hours_50' THEN 50
	WHEN code = 'hours_100' THEN 100
	ELSE NULL
END, points_reward
FROM achievements
ON CONFLICT (code) DO UPDATE
SET threshold = EXCLUDED.threshold,
	points = EXCLUDED.points;

CREATE OR REPLACE FUNCTION recalculate_user_gamification(target_user_id UUID)
RETURNS void
LANGUAGE plpgsql
AS $$
DECLARE
	total_points INTEGER;
BEGIN
	INSERT INTO volunteer_profiles (user_id)
	VALUES (target_user_id)
	ON CONFLICT (user_id) DO NOTHING;

	WITH eligible AS (
		SELECT a.id, a.points_reward, a.name
		FROM achievements a
		JOIN volunteer_profiles vp ON vp.user_id = target_user_id
		JOIN users u ON u.id = target_user_id
		WHERE
			(a.code = 'registration')
			OR (
				a.code = 'profile_completed'
				AND (
					COALESCE(vp.bio, '') <> ''
					OR COALESCE(vp.city, '') <> ''
					OR COALESCE(vp.phone, '') <> ''
					OR COALESCE(vp.interests, '') <> ''
				)
			)
			OR (
				a.code = 'first_participation'
				AND EXISTS (
					SELECT 1 FROM event_attendance ea WHERE ea.user_id = target_user_id
				)
			)
			OR (
				a.code = 'task_completed'
				AND EXISTS (
					SELECT 1
					FROM task_assignments ta
					JOIN tasks t ON t.id = ta.task_id
					WHERE ta.user_id = target_user_id AND t.status = 'completed'
				)
			)
			OR (a.code = 'hours_10' AND vp.total_hours >= 10)
			OR (a.code = 'hours_25' AND vp.total_hours >= 25)
			OR (a.code = 'hours_50' AND vp.total_hours >= 50)
			OR (a.code = 'hours_100' AND vp.total_hours >= 100)
	),
	new_awards AS (
		INSERT INTO volunteer_achievements (user_id, achievement_id)
		SELECT target_user_id, id FROM eligible
		ON CONFLICT (user_id, achievement_id) DO NOTHING
		RETURNING achievement_id
	)
	INSERT INTO points_transactions (user_id, achievement_id, points, reason)
	SELECT target_user_id, a.id, a.points_reward, a.name
	FROM achievements a
	JOIN new_awards na ON na.achievement_id = a.id
	WHERE a.points_reward <> 0
	ON CONFLICT (user_id, achievement_id, source_type) DO NOTHING;

	SELECT COALESCE(SUM(a.points_reward), 0)
	INTO total_points
	FROM volunteer_achievements va
	JOIN achievements a ON a.id = va.achievement_id
	WHERE va.user_id = target_user_id;

	UPDATE volunteer_profiles
	SET points = total_points,
		level = GREATEST(1, FLOOR(total_points / 100.0)::INTEGER + 1),
		updated_at = now()
	WHERE user_id = target_user_id;
END;
$$;

SELECT recalculate_user_gamification(id) FROM users;

CREATE INDEX IF NOT EXISTS idx_volunteer_achievements_user_id ON volunteer_achievements(user_id);
CREATE INDEX IF NOT EXISTS idx_points_transactions_user_id ON points_transactions(user_id);
