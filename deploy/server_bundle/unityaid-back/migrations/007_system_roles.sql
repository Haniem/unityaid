CREATE TABLE IF NOT EXISTS system_roles (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	code TEXT NOT NULL UNIQUE,
	name TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS user_system_roles (
	user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	role_id UUID NOT NULL REFERENCES system_roles(id) ON DELETE CASCADE,
	assigned_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	PRIMARY KEY (user_id, role_id)
);

INSERT INTO system_roles (code, name, description)
VALUES
	('system_admin', 'Системный администратор', 'Полный доступ ко всей системе, настройкам, пользователям, организациям, справочникам и аудиту.'),
	('system_manager', 'Системный менеджер', 'Управление пользователями, справочниками и проверка данных без критичных системных настроек.'),
	('support', 'Поддержка', 'Просмотр данных и помощь пользователям без права менять роли и удалять данные.'),
	('user', 'Пользователь', 'Обычный пользователь системы, права которого дальше определяются членством в организациях.')
ON CONFLICT (code) DO UPDATE SET
	name = EXCLUDED.name,
	description = EXCLUDED.description;

INSERT INTO user_system_roles (user_id, role_id)
SELECT DISTINCT om.user_id, sr.id
FROM organization_members om
JOIN system_roles sr ON sr.code = 'system_admin'
WHERE om.role = 'super_admin'
ON CONFLICT DO NOTHING;

INSERT INTO user_system_roles (user_id, role_id)
SELECT u.id, sr.id
FROM users u
JOIN system_roles sr ON sr.code = 'user'
ON CONFLICT DO NOTHING;
