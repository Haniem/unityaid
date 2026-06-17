INSERT INTO organizations (id, name, slug, description, contact_email)
VALUES
	('11111111-1111-1111-1111-111111111111', 'Пульс Добра', 'puls-dobra', 'Демо-организация для защиты диплома: городские волонтерские акции, задачи, часы, новости и аналитика.', 'team@puls.test')
ON CONFLICT (id) DO UPDATE SET
	name = EXCLUDED.name,
	slug = EXCLUDED.slug,
	description = EXCLUDED.description,
	contact_email = EXCLUDED.contact_email,
	updated_at = now();

-- Пароль для всех демо-пользователей: password
INSERT INTO users (id, email, password_hash, first_name, last_name, patronymic, avatar_url, locale, is_active, is_email_verified)
VALUES
	('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'admin@puls.test', '$2a$10$uISbNqAL/.aA87KggeAr6OohJh9J.uY2i7O8k7qJ5WqsPBFO9/ARa', 'Павел', 'Зозин', 'Алексеевич', 'https://i.pravatar.cc/160?img=12', 'ru', true, true),
	('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'admin1@puls.test', '$2a$10$uISbNqAL/.aA87KggeAr6OohJh9J.uY2i7O8k7qJ5WqsPBFO9/ARa', 'Анна', 'Крылова', 'Игоревна', 'https://i.pravatar.cc/160?img=5', 'ru', true, true),
	('cccccccc-cccc-cccc-cccc-cccccccccccc', 'admin2@puls.test', '$2a$10$uISbNqAL/.aA87KggeAr6OohJh9J.uY2i7O8k7qJ5WqsPBFO9/ARa', 'Мария', 'Соколова', 'Петровна', 'https://i.pravatar.cc/160?img=9', 'ru', true, true),
	('dddddddd-dddd-dddd-dddd-dddddddddddd', 'admin3@puls.test', '$2a$10$uISbNqAL/.aA87KggeAr6OohJh9J.uY2i7O8k7qJ5WqsPBFO9/ARa', 'Илья', 'Морозов', NULL, 'https://i.pravatar.cc/160?img=15', 'ru', true, true)
ON CONFLICT (id) DO UPDATE SET
	email = EXCLUDED.email,
	password_hash = EXCLUDED.password_hash,
	first_name = EXCLUDED.first_name,
	last_name = EXCLUDED.last_name,
	patronymic = EXCLUDED.patronymic,
	avatar_url = EXCLUDED.avatar_url,
	locale = EXCLUDED.locale,
	is_active = EXCLUDED.is_active,
	is_email_verified = EXCLUDED.is_email_verified,
	updated_at = now();

INSERT INTO organization_members (organization_id, user_id, role, status)
VALUES
	('11111111-1111-1111-1111-111111111111', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'super_admin', 'active'),
	('11111111-1111-1111-1111-111111111111', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'org_admin', 'active'),
	('11111111-1111-1111-1111-111111111111', 'cccccccc-cccc-cccc-cccc-cccccccccccc', 'coordinator', 'active'),
	('11111111-1111-1111-1111-111111111111', 'dddddddd-dddd-dddd-dddd-dddddddddddd', 'volunteer', 'active')
ON CONFLICT (organization_id, user_id) DO UPDATE SET
	role = EXCLUDED.role,
	status = EXCLUDED.status,
	updated_at = now();
