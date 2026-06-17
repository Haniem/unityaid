-- Демо-сценарий для защиты: 4 основных пользователя, пароль у всех: password.

DELETE FROM users
WHERE id IN (
	'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee',
	'ffffffff-ffff-ffff-ffff-ffffffffffff',
	'99999999-9999-9999-9999-999999999999',
	'88888888-8888-8888-8888-888888888888'
);

DELETE FROM organizations
WHERE id = '22222222-2222-2222-2222-222222222222';

INSERT INTO organizations (id, name, slug, description, contact_email)
VALUES
	(
		'11111111-1111-1111-1111-111111111111',
		'Пульс Добра',
		'puls-dobra',
		'Демо-организация для защиты диплома: городские волонтерские акции, задачи, часы, новости и аналитика.',
		'team@puls.test'
	)
ON CONFLICT (id) DO UPDATE SET
	name = EXCLUDED.name,
	slug = EXCLUDED.slug,
	description = EXCLUDED.description,
	contact_email = EXCLUDED.contact_email,
	updated_at = now();

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

INSERT INTO volunteer_profiles (user_id, city, phone, bio, total_hours, points, level)
VALUES
	('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Екатеринбург', '+7 900 000-10-00', 'Главный администратор демо-стенда. Проверяет настройки, аналитику и права доступа.', 0, 0, 1),
	('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'Екатеринбург', '+7 900 000-10-01', 'Администратор организации: ведет команду, новости и структуру волонтерского центра.', 6, 80, 1),
	('cccccccc-cccc-cccc-cccc-cccccccccccc', 'Екатеринбург', '+7 900 000-10-02', 'Менеджер организации: планирует мероприятия, назначает задачи и подтверждает часы.', 12.5, 170, 2),
	('dddddddd-dddd-dddd-dddd-dddddddddddd', 'Екатеринбург', '+7 900 000-10-03', 'Волонтер: участвует в акциях, выполняет задачи, получает достижения и сертификаты.', 28.5, 360, 4)
ON CONFLICT (user_id) DO UPDATE SET
	city = EXCLUDED.city,
	phone = EXCLUDED.phone,
	bio = EXCLUDED.bio,
	total_hours = EXCLUDED.total_hours,
	points = EXCLUDED.points,
	level = EXCLUDED.level,
	updated_at = now();

INSERT INTO skills (name)
VALUES ('Коммуникация'), ('Первая помощь'), ('Логистика'), ('Фото и видео'), ('Работа с детьми'), ('Координация смен')
ON CONFLICT (name) DO NOTHING;

INSERT INTO volunteer_skills (volunteer_profile_id, skill_id)
SELECT vp.id, s.id
FROM volunteer_profiles vp
JOIN users u ON u.id = vp.user_id
JOIN skills s ON s.name IN ('Коммуникация', 'Логистика', 'Работа с детьми')
WHERE u.email = 'admin3@puls.test'
ON CONFLICT DO NOTHING;

INSERT INTO events (id, organization_id, title, description, format, status, starts_at, ends_at, location, max_participants, created_by)
VALUES
	(
		'33333333-3333-3333-3333-333333333333',
		'11111111-1111-1111-1111-111111111111',
		'Сбор гуманитарной помощи',
		'Прием, сортировка и упаковка вещей для семей в трудной ситуации. Основное событие для показа заявок, смен и задач.',
		'offline',
		'published',
		now() + interval '4 days',
		now() + interval '4 days 4 hours',
		'ул. Мира, 10',
		25,
		'cccccccc-cccc-cccc-cccc-cccccccccccc'
	),
	(
		'44444444-4444-4444-4444-444444444444',
		'11111111-1111-1111-1111-111111111111',
		'Онлайн-инструктаж для новичков',
		'Короткая встреча о правилах участия, личном кабинете, отметке часов и получении сертификатов.',
		'online',
		'published',
		now() + interval '2 days',
		now() + interval '2 days 1 hour',
		'Zoom',
		80,
		'cccccccc-cccc-cccc-cccc-cccccccccccc'
	),
	(
		'55555555-5555-5555-5555-555555555555',
		'11111111-1111-1111-1111-111111111111',
		'Городской форум добровольцев',
		'Завершенное мероприятие для демонстрации посещаемости, подтвержденных часов, достижений и сертификата.',
		'offline',
		'completed',
		now() - interval '10 days',
		now() - interval '10 days' + interval '5 hours',
		'Дом молодежи',
		60,
		'cccccccc-cccc-cccc-cccc-cccccccccccc'
	)
ON CONFLICT (id) DO UPDATE SET
	organization_id = EXCLUDED.organization_id,
	title = EXCLUDED.title,
	description = EXCLUDED.description,
	format = EXCLUDED.format,
	status = EXCLUDED.status,
	starts_at = EXCLUDED.starts_at,
	ends_at = EXCLUDED.ends_at,
	location = EXCLUDED.location,
	max_participants = EXCLUDED.max_participants,
	created_by = EXCLUDED.created_by,
	updated_at = now();

INSERT INTO event_shifts (id, event_id, title, starts_at, ends_at, capacity)
VALUES
	('18181818-1818-1818-1818-181818181818', '33333333-3333-3333-3333-333333333333', 'Прием вещей', now() + interval '4 days', now() + interval '4 days 2 hours', 12),
	('19191919-1919-1919-1919-191919191919', '33333333-3333-3333-3333-333333333333', 'Сортировка и упаковка', now() + interval '4 days 2 hours', now() + interval '4 days 4 hours', 14),
	('1a1a1a1a-1a1a-1a1a-1a1a-1a1a1a1a1a1a', '55555555-5555-5555-5555-555555555555', 'Регистрация участников', now() - interval '10 days', now() - interval '10 days' + interval '5 hours', 8)
ON CONFLICT (id) DO NOTHING;

INSERT INTO event_applications (id, event_id, user_id, status, message, created_at, updated_at)
VALUES
	('1b1b1b1b-1b1b-1b1b-1b1b-1b1b1b1b1b1b', '33333333-3333-3333-3333-333333333333', 'dddddddd-dddd-dddd-dddd-dddddddddddd', 'approved', 'Готов выйти на смену сортировки.', now() - interval '1 day', now() - interval '1 day'),
	('1c1c1c1c-1c1c-1c1c-1c1c-1c1c1c1c1c1c', '44444444-4444-4444-4444-444444444444', 'dddddddd-dddd-dddd-dddd-dddddddddddd', 'pending', 'Хочу пройти инструктаж перед мероприятием.', now() - interval '3 hours', now() - interval '3 hours'),
	('1d1d1d1d-1d1d-1d1d-1d1d-1d1d1d1d1d1d', '55555555-5555-5555-5555-555555555555', 'dddddddd-dddd-dddd-dddd-dddddddddddd', 'approved', 'Участвовал в регистрации гостей.', now() - interval '12 days', now() - interval '11 days')
ON CONFLICT (event_id, user_id) DO UPDATE SET
	status = EXCLUDED.status,
	message = EXCLUDED.message,
	updated_at = EXCLUDED.updated_at;

INSERT INTO event_attendance (id, event_id, user_id, check_in_at, check_out_at, hours)
VALUES
	('1e1e1e1e-1e1e-1e1e-1e1e-1e1e1e1e1e1e', '55555555-5555-5555-5555-555555555555', 'dddddddd-dddd-dddd-dddd-dddddddddddd', now() - interval '10 days', now() - interval '10 days' + interval '5 hours', 5)
ON CONFLICT (event_id, user_id) DO UPDATE SET
	check_in_at = EXCLUDED.check_in_at,
	check_out_at = EXCLUDED.check_out_at,
	hours = EXCLUDED.hours,
	updated_at = now();

INSERT INTO event_feedback (id, event_id, user_id, rating, comment)
VALUES
	('1f1f1f1f-1f1f-1f1f-1f1f-1f1f1f1f1f1f', '55555555-5555-5555-5555-555555555555', 'dddddddd-dddd-dddd-dddd-dddddddddddd', 5, 'Понятная организация и быстрый чек-ин через систему.')
ON CONFLICT (event_id, user_id) DO UPDATE SET
	rating = EXCLUDED.rating,
	comment = EXCLUDED.comment;

INSERT INTO tasks (id, organization_id, event_id, title, description, status, priority, due_at, created_by, completion_confirmed_by, completion_confirmed_at)
VALUES
	(
		'66666666-6666-6666-6666-666666666666',
		'11111111-1111-1111-1111-111111111111',
		'33333333-3333-3333-3333-333333333333',
		'Подготовить зону сортировки',
		'Разметить столы, подготовить коробки и навигационные таблички.',
		'assigned',
		'high',
		now() + interval '3 days',
		'cccccccc-cccc-cccc-cccc-cccccccccccc',
		NULL,
		NULL
	),
	(
		'77777777-7777-7777-7777-777777777777',
		'11111111-1111-1111-1111-111111111111',
		'33333333-3333-3333-3333-333333333333',
		'Согласовать список волонтеров',
		'Проверить заявки, подтвердить участников и отправить памятку перед событием.',
		'in_progress',
		'medium',
		now() + interval '2 days',
		'cccccccc-cccc-cccc-cccc-cccccccccccc',
		NULL,
		NULL
	),
	(
		'78787878-7878-7878-7878-787878787878',
		'11111111-1111-1111-1111-111111111111',
		'55555555-5555-5555-5555-555555555555',
		'Подготовить отчет по форуму',
		'Собрать посещаемость, часы и короткий отчет для администратора организации.',
		'completed',
		'medium',
		now() - interval '8 days',
		'cccccccc-cccc-cccc-cccc-cccccccccccc',
		'cccccccc-cccc-cccc-cccc-cccccccccccc',
		now() - interval '8 days'
	)
ON CONFLICT (id) DO UPDATE SET
	organization_id = EXCLUDED.organization_id,
	event_id = EXCLUDED.event_id,
	title = EXCLUDED.title,
	description = EXCLUDED.description,
	status = EXCLUDED.status,
	priority = EXCLUDED.priority,
	due_at = EXCLUDED.due_at,
	created_by = EXCLUDED.created_by,
	completion_confirmed_by = EXCLUDED.completion_confirmed_by,
	completion_confirmed_at = EXCLUDED.completion_confirmed_at,
	updated_at = now();

INSERT INTO task_assignments (task_id, user_id, role)
VALUES
	('66666666-6666-6666-6666-666666666666', 'dddddddd-dddd-dddd-dddd-dddddddddddd', 'assignee'),
	('77777777-7777-7777-7777-777777777777', 'cccccccc-cccc-cccc-cccc-cccccccccccc', 'owner'),
	('78787878-7878-7878-7878-787878787878', 'dddddddd-dddd-dddd-dddd-dddddddddddd', 'assignee')
ON CONFLICT (task_id, user_id) DO UPDATE SET
	role = EXCLUDED.role;

INSERT INTO task_comments (id, task_id, user_id, content)
VALUES
	('20202020-2020-2020-2020-202020202020', '66666666-6666-6666-6666-666666666666', 'cccccccc-cccc-cccc-cccc-cccccccccccc', 'Илья, возьми зону сортировки. Таблички и коробки будут у входа.'),
	('21212121-2121-2121-2121-212121212121', '66666666-6666-6666-6666-666666666666', 'dddddddd-dddd-dddd-dddd-dddddddddddd', 'Принял, подготовлю схему расстановки заранее.')
ON CONFLICT (id) DO UPDATE SET
	content = EXCLUDED.content;

INSERT INTO task_status_history (id, task_id, from_status, to_status, changed_by, created_at)
VALUES
	('22222222-3333-4444-5555-666666666666', '66666666-6666-6666-6666-666666666666', 'created', 'assigned', 'cccccccc-cccc-cccc-cccc-cccccccccccc', now() - interval '12 hours'),
	('23232323-2323-2323-2323-232323232323', '78787878-7878-7878-7878-787878787878', 'review', 'completed', 'cccccccc-cccc-cccc-cccc-cccccccccccc', now() - interval '8 days')
ON CONFLICT (id) DO UPDATE SET
	from_status = EXCLUDED.from_status,
	to_status = EXCLUDED.to_status,
	changed_by = EXCLUDED.changed_by,
	created_at = EXCLUDED.created_at;

INSERT INTO time_entries (id, organization_id, user_id, event_id, task_id, hours, description, status, reviewed_by, reviewed_at, created_at, updated_at)
VALUES
	('24242424-2424-2424-2424-242424242424', '11111111-1111-1111-1111-111111111111', 'dddddddd-dddd-dddd-dddd-dddddddddddd', '55555555-5555-5555-5555-555555555555', NULL, 5, 'Регистрация участников городского форума.', 'approved', 'cccccccc-cccc-cccc-cccc-cccccccccccc', now() - interval '9 days', now() - interval '10 days', now() - interval '9 days'),
	('25252525-2525-2525-2525-252525252525', '11111111-1111-1111-1111-111111111111', 'dddddddd-dddd-dddd-dddd-dddddddddddd', NULL, '78787878-7878-7878-7878-787878787878', 3.5, 'Подготовил отчет и список участников после форума.', 'approved', 'cccccccc-cccc-cccc-cccc-cccccccccccc', now() - interval '8 days', now() - interval '8 days', now() - interval '8 days'),
	('26262626-2626-2626-2626-262626262626', '11111111-1111-1111-1111-111111111111', 'dddddddd-dddd-dddd-dddd-dddddddddddd', NULL, '66666666-6666-6666-6666-666666666666', 2, 'Подготовка схемы зоны сортировки.', 'pending', NULL, NULL, now() - interval '4 hours', now() - interval '4 hours')
ON CONFLICT (id) DO UPDATE SET
	hours = EXCLUDED.hours,
	description = EXCLUDED.description,
	status = EXCLUDED.status,
	reviewed_by = EXCLUDED.reviewed_by,
	reviewed_at = EXCLUDED.reviewed_at,
	updated_at = EXCLUDED.updated_at;

INSERT INTO certificates (id, user_id, organization_id, type, title, description, total_hours, verify_code, issued_by, issued_at)
VALUES
	(
		'27272727-2727-2727-2727-272727272727',
		'dddddddd-dddd-dddd-dddd-dddddddddddd',
		'11111111-1111-1111-1111-111111111111',
		'participation',
		'Сертификат волонтера форума',
		'Подтверждает участие в городском форуме добровольцев и помощь на регистрации участников.',
		8.5,
		'PULS-DEMO-2026',
		'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
		now() - interval '7 days'
	)
ON CONFLICT (verify_code) DO UPDATE SET
	user_id = EXCLUDED.user_id,
	organization_id = EXCLUDED.organization_id,
	title = EXCLUDED.title,
	description = EXCLUDED.description,
	total_hours = EXCLUDED.total_hours,
	issued_by = EXCLUDED.issued_by,
	issued_at = EXCLUDED.issued_at;

SELECT recalculate_user_gamification(id)
FROM users
WHERE email IN ('admin@puls.test', 'admin1@puls.test', 'admin2@puls.test', 'admin3@puls.test');

UPDATE notifications
SET
	is_read = true,
	read_at = now() - interval '2 days',
	created_at = now() - interval '2 days'
WHERE type = 'achievement_earned'
	AND user_id IN (
		'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
		'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
		'cccccccc-cccc-cccc-cccc-cccccccccccc',
		'dddddddd-dddd-dddd-dddd-dddddddddddd'
	);

INSERT INTO notifications (user_id, type, title, body, link, entity_type, entity_id, is_read, read_at, created_at)
VALUES
	('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'demo_admin_overview', 'Готова сводка по демо-стенду', 'Проверьте аналитику, роли и состояние ключевых модулей перед защитой.', '/analytics', 'demo', 'admin-overview', false, NULL, now() - interval '25 minutes'),
	('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'demo_settings', 'Первичная настройка заполнена', 'Организация, роли и демо-пользователи готовы к показу.', '/settings', 'demo', 'settings', true, now() - interval '1 hour', now() - interval '1 hour'),
	('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'demo_org_review', 'Новая заявка ожидает контроля', 'Менеджер согласовал волонтера на сбор гуманитарной помощи.', '/calendar/33333333-3333-3333-3333-333333333333', 'event', '33333333-3333-3333-3333-333333333333', false, NULL, now() - interval '18 minutes'),
	('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'demo_news_ready', 'Новость опубликована', 'Анонс сбора гуманитарной помощи доступен в ленте.', '/news/12121212-1212-1212-1212-121212121212', 'news', '12121212-1212-1212-1212-121212121212', true, now() - interval '2 hours', now() - interval '2 hours'),
	('cccccccc-cccc-cccc-cccc-cccccccccccc', 'demo_task_due', 'Задача требует внимания', 'До согласования списка волонтеров осталось меньше двух дней.', '/tasks/77777777-7777-7777-7777-777777777777', 'task', '77777777-7777-7777-7777-777777777777', false, NULL, now() - interval '12 minutes'),
	('cccccccc-cccc-cccc-cccc-cccccccccccc', 'demo_hours_pending', 'Часы волонтера ожидают проверки', 'Илья добавил 2 часа по подготовке зоны сортировки.', '/time-entries', 'time_entry', '26262626-2626-2626-2626-262626262626', false, NULL, now() - interval '34 minutes'),
	('dddddddd-dddd-dddd-dddd-dddddddddddd', 'demo_event_approved', 'Заявка одобрена', 'Вы включены в команду сбора гуманитарной помощи.', '/calendar/33333333-3333-3333-3333-333333333333', 'event', '33333333-3333-3333-3333-333333333333', false, NULL, now() - interval '8 minutes'),
	('dddddddd-dddd-dddd-dddd-dddddddddddd', 'demo_certificate', 'Сертификат доступен', 'Сертификат за городской форум можно скачать в профиле.', '/certificates', 'certificate', '27272727-2727-2727-2727-272727272727', true, now() - interval '1 day', now() - interval '1 day')
ON CONFLICT (user_id, type, entity_type, entity_id) DO UPDATE SET
	title = EXCLUDED.title,
	body = EXCLUDED.body,
	link = EXCLUDED.link,
	is_read = EXCLUDED.is_read,
	read_at = EXCLUDED.read_at,
	created_at = EXCLUDED.created_at;
