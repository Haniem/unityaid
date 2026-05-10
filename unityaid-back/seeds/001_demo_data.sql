INSERT INTO organizations (id, name, slug, description, contact_email)
VALUES
	('11111111-1111-1111-1111-111111111111', 'Добрые руки', 'dobrye-ruki', 'Городская волонтерская организация помощи людям и социальным учреждениям.', 'hello@dobrye-ruki.test'),
	('22222222-2222-2222-2222-222222222222', 'ЭкоПульс', 'ecopulse', 'Экологическое движение: уборки, лекции, раздельный сбор и городские акции.', 'team@ecopulse.test')
ON CONFLICT (id) DO NOTHING;

-- Пароль для всех демо-пользователей: password
INSERT INTO users (id, email, password_hash, first_name, last_name, patronymic, avatar_url, locale)
VALUES
	('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'admin@unityaid.test', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'Павел', 'Зозин', 'Алексеевич', 'https://i.pravatar.cc/160?img=12', 'ru'),
	('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'org.admin@dobrye-ruki.test', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'Анна', 'Крылова', 'Игоревна', 'https://i.pravatar.cc/160?img=5', 'ru'),
	('cccccccc-cccc-cccc-cccc-cccccccccccc', 'coord@dobrye-ruki.test', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'Мария', 'Соколова', 'Петровна', 'https://i.pravatar.cc/160?img=9', 'ru'),
	('dddddddd-dddd-dddd-dddd-dddddddddddd', 'volunteer1@test.local', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'Илья', 'Морозов', NULL, 'https://i.pravatar.cc/160?img=15', 'ru'),
	('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'volunteer2@test.local', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'Екатерина', 'Волкова', NULL, 'https://i.pravatar.cc/160?img=20', 'ru'),
	('ffffffff-ffff-ffff-ffff-ffffffffffff', 'org.admin@ecopulse.test', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'Олег', 'Никитин', NULL, 'https://i.pravatar.cc/160?img=32', 'ru'),
	('99999999-9999-9999-9999-999999999999', 'coord@ecopulse.test', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'Наталья', 'Егорова', NULL, 'https://i.pravatar.cc/160?img=47', 'ru'),
	('88888888-8888-8888-8888-888888888888', 'volunteer3@test.local', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'Денис', 'Лебедев', NULL, 'https://i.pravatar.cc/160?img=57', 'ru')
ON CONFLICT (id) DO NOTHING;

INSERT INTO organization_members (organization_id, user_id, role)
VALUES
	('11111111-1111-1111-1111-111111111111', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'super_admin'),
	('22222222-2222-2222-2222-222222222222', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'super_admin'),
	('11111111-1111-1111-1111-111111111111', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'org_admin'),
	('11111111-1111-1111-1111-111111111111', 'cccccccc-cccc-cccc-cccc-cccccccccccc', 'coordinator'),
	('11111111-1111-1111-1111-111111111111', 'dddddddd-dddd-dddd-dddd-dddddddddddd', 'volunteer'),
	('11111111-1111-1111-1111-111111111111', 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'volunteer'),
	('22222222-2222-2222-2222-222222222222', 'ffffffff-ffff-ffff-ffff-ffffffffffff', 'org_admin'),
	('22222222-2222-2222-2222-222222222222', '99999999-9999-9999-9999-999999999999', 'coordinator'),
	('22222222-2222-2222-2222-222222222222', '88888888-8888-8888-8888-888888888888', 'volunteer')
ON CONFLICT (organization_id, user_id) DO NOTHING;

INSERT INTO volunteer_profiles (user_id, city, phone, bio, total_hours, points, level)
VALUES
	('dddddddd-dddd-dddd-dddd-dddddddddddd', 'Екатеринбург', '+7 900 000-10-01', 'Помогаю на городских событиях и социальных акциях.', 18.5, 240, 3),
	('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'Екатеринбург', '+7 900 000-10-02', 'Интересуюсь образовательными и благотворительными проектами.', 9, 120, 2),
	('88888888-8888-8888-8888-888888888888', 'Екатеринбург', '+7 900 000-10-03', 'Участвую в экологических акциях и сортировке отходов.', 14, 180, 2)
ON CONFLICT (user_id) DO NOTHING;

INSERT INTO skills (name)
VALUES ('Коммуникация'), ('Первая помощь'), ('Логистика'), ('Фото и видео'), ('Экология'), ('Работа с детьми')
ON CONFLICT (name) DO NOTHING;

INSERT INTO events (id, organization_id, title, description, format, status, starts_at, ends_at, location, max_participants, created_by)
VALUES
	('33333333-3333-3333-3333-333333333333', '11111111-1111-1111-1111-111111111111', 'Сбор гуманитарной помощи', 'Прием, сортировка и упаковка вещей для семей в трудной ситуации.', 'offline', 'published', now() + interval '4 days', now() + interval '4 days 4 hours', 'ул. Мира, 10', 25, 'cccccccc-cccc-cccc-cccc-cccccccccccc'),
	('44444444-4444-4444-4444-444444444444', '11111111-1111-1111-1111-111111111111', 'Онлайн-инструктаж для новичков', 'Знакомство с правилами участия и базовыми процессами организации.', 'online', 'published', now() + interval '2 days', now() + interval '2 days 1 hour', 'Zoom', 80, 'cccccccc-cccc-cccc-cccc-cccccccccccc'),
	('55555555-5555-5555-5555-555555555555', '22222222-2222-2222-2222-222222222222', 'Чистый берег', 'Городская экологическая акция по уборке береговой линии.', 'offline', 'published', now() + interval '7 days', now() + interval '7 days 5 hours', 'Парк у набережной', 40, '99999999-9999-9999-9999-999999999999')
ON CONFLICT (id) DO NOTHING;

INSERT INTO tasks (id, organization_id, event_id, title, description, status, priority, due_at, created_by)
VALUES
	('66666666-6666-6666-6666-666666666666', '11111111-1111-1111-1111-111111111111', '33333333-3333-3333-3333-333333333333', 'Подготовить зону сортировки', 'Разметить столы, подготовить коробки и навигационные таблички.', 'assigned', 'high', now() + interval '3 days', 'cccccccc-cccc-cccc-cccc-cccccccccccc'),
	('77777777-7777-7777-7777-777777777777', '22222222-2222-2222-2222-222222222222', '55555555-5555-5555-5555-555555555555', 'Проверить инвентарь', 'Сверить перчатки, мешки и аптечку перед акцией.', 'created', 'medium', now() + interval '6 days', '99999999-9999-9999-9999-999999999999')
ON CONFLICT (id) DO NOTHING;

INSERT INTO task_assignments (task_id, user_id)
VALUES ('66666666-6666-6666-6666-666666666666', 'dddddddd-dddd-dddd-dddd-dddddddddddd')
ON CONFLICT (task_id, user_id) DO NOTHING;
