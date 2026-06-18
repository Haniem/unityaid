INSERT INTO events (id, organization_id, title, description, format, status, starts_at, ends_at, location, max_participants, created_by)
VALUES
	('6a6a6a6a-6a6a-6a6a-6a6a-6a6a6a6a6a6a', '11111111-1111-1111-1111-111111111111', 'Экологическая акция у набережной', 'Сбор мусора, сортировка вторсырья и фиксация результата для отчета организации.', 'offline', 'completed', now() - interval '18 days', now() - interval '18 days' + interval '4 hours', 'Набережная городского пруда', 35, 'cccccccc-cccc-cccc-cccc-cccccccccccc'),
	('6b6b6b6b-6b6b-6b6b-6b6b-6b6b6b6b6b6b', '11111111-1111-1111-1111-111111111111', 'Сбор гуманитарных наборов', 'Подготовка продуктовых наборов и адресная передача заявок координатору.', 'offline', 'published', now() + interval '6 days', now() + interval '6 days 3 hours', 'Волонтерский штаб', 24, 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb'),
	('6c6c6c6c-6c6c-6c6c-6c6c-6c6c6c6c6c6c', '11111111-1111-1111-1111-111111111111', 'Онлайн-инструктаж волонтеров', 'Короткая подготовка участников перед выездными сменами и проверка готовности команд.', 'online', 'completed', now() - interval '5 days', now() - interval '5 days' + interval '1 hour 30 minutes', 'Онлайн', 60, 'cccccccc-cccc-cccc-cccc-cccccccccccc')
ON CONFLICT (id) DO UPDATE SET
	title = EXCLUDED.title,
	description = EXCLUDED.description,
	format = EXCLUDED.format,
	status = EXCLUDED.status,
	starts_at = EXCLUDED.starts_at,
	ends_at = EXCLUDED.ends_at,
	location = EXCLUDED.location,
	max_participants = EXCLUDED.max_participants,
	updated_at = now();

INSERT INTO event_applications (id, event_id, user_id, status, message, created_at, updated_at)
VALUES
	('6d6d6d6d-6d6d-6d6d-6d6d-6d6d6d6d6d6d', '6a6a6a6a-6a6a-6a6a-6a6a-6a6a6a6a6a6a', 'dddddddd-dddd-dddd-dddd-dddddddddddd', 'approved', 'Готов помогать на зоне сортировки.', now() - interval '20 days', now() - interval '19 days'),
	('6e6e6e6e-6e6e-6e6e-6e6e-6e6e6e6e6e6e', '6b6b6b6b-6b6b-6b6b-6b6b-6b6b6b6b6b6b', 'dddddddd-dddd-dddd-dddd-dddddddddddd', 'pending', 'Могу выйти на вечернюю смену.', now() - interval '2 hours', now() - interval '2 hours'),
	('6f6f6f6f-6f6f-6f6f-6f6f-6f6f6f6f6f6f', '6c6c6c6c-6c6c-6c6c-6c6c-6c6c6c6c6c6c', 'dddddddd-dddd-dddd-dddd-dddddddddddd', 'approved', 'Нужно пройти инструктаж перед мероприятием.', now() - interval '7 days', now() - interval '7 days'),
	('70707070-7070-7070-7070-707070707070', '6b6b6b6b-6b6b-6b6b-6b6b-6b6b6b6b6b6b', 'cccccccc-cccc-cccc-cccc-cccccccccccc', 'approved', 'Возьму координацию упаковки.', now() - interval '1 day', now() - interval '1 day')
ON CONFLICT (event_id, user_id) DO UPDATE SET
	status = EXCLUDED.status,
	message = EXCLUDED.message,
	updated_at = EXCLUDED.updated_at;

INSERT INTO event_attendance (id, event_id, user_id, check_in_at, check_out_at, hours)
VALUES
	('71717171-7171-7171-7171-717171717171', '6a6a6a6a-6a6a-6a6a-6a6a-6a6a6a6a6a6a', 'dddddddd-dddd-dddd-dddd-dddddddddddd', now() - interval '18 days', now() - interval '18 days' + interval '4 hours', 4),
	('72727272-7272-7272-7272-727272727272', '6c6c6c6c-6c6c-6c6c-6c6c-6c6c6c6c6c6c', 'dddddddd-dddd-dddd-dddd-dddddddddddd', now() - interval '5 days', now() - interval '5 days' + interval '1 hour 30 minutes', 1.5)
ON CONFLICT (event_id, user_id) DO UPDATE SET
	check_in_at = EXCLUDED.check_in_at,
	check_out_at = EXCLUDED.check_out_at,
	hours = EXCLUDED.hours,
	updated_at = now();

INSERT INTO tasks (id, organization_id, event_id, title, description, status, priority, due_at, created_by, completion_confirmed_by, completion_confirmed_at)
VALUES
	('73737373-7373-7373-7373-737373737373', '11111111-1111-1111-1111-111111111111', '6b6b6b6b-6b6b-6b6b-6b6b-6b6b6b6b6b6b', 'Собрать список получателей наборов', 'Проверить заявки, исключить дубли и передать финальный список координатору.', 'in_progress', 'high', now() + interval '3 days', 'cccccccc-cccc-cccc-cccc-cccccccccccc', NULL, NULL),
	('74747474-7474-7474-7474-747474747474', '11111111-1111-1111-1111-111111111111', '6b6b6b6b-6b6b-6b6b-6b6b-6b6b6b6b6b6b', 'Подготовить зону выдачи', 'Разметить столы, подготовить бейджи и коробки по категориям.', 'assigned', 'medium', now() + interval '5 days', 'cccccccc-cccc-cccc-cccc-cccccccccccc', NULL, NULL),
	('75757575-7575-7575-7575-757575757575', '11111111-1111-1111-1111-111111111111', '6a6a6a6a-6a6a-6a6a-6a6a-6a6a6a6a6a6a', 'Загрузить фотоотчет по акции', 'Добавить итоговые фотографии и краткое описание результата в отчет.', 'completed', 'low', now() - interval '17 days', 'cccccccc-cccc-cccc-cccc-cccccccccccc', 'cccccccc-cccc-cccc-cccc-cccccccccccc', now() - interval '17 days')
ON CONFLICT (id) DO UPDATE SET
	title = EXCLUDED.title,
	description = EXCLUDED.description,
	status = EXCLUDED.status,
	priority = EXCLUDED.priority,
	due_at = EXCLUDED.due_at,
	completion_confirmed_by = EXCLUDED.completion_confirmed_by,
	completion_confirmed_at = EXCLUDED.completion_confirmed_at,
	updated_at = now();

INSERT INTO task_assignments (task_id, user_id, role)
VALUES
	('73737373-7373-7373-7373-737373737373', 'cccccccc-cccc-cccc-cccc-cccccccccccc', 'owner'),
	('73737373-7373-7373-7373-737373737373', 'dddddddd-dddd-dddd-dddd-dddddddddddd', 'assignee'),
	('74747474-7474-7474-7474-747474747474', 'dddddddd-dddd-dddd-dddd-dddddddddddd', 'assignee'),
	('75757575-7575-7575-7575-757575757575', 'dddddddd-dddd-dddd-dddd-dddddddddddd', 'assignee')
ON CONFLICT (task_id, user_id) DO UPDATE SET
	role = EXCLUDED.role;

INSERT INTO time_entries (id, organization_id, user_id, event_id, task_id, hours, description, status, reviewed_by, reviewed_at, created_at, updated_at)
VALUES
	('76767676-7676-7676-7676-767676767676', '11111111-1111-1111-1111-111111111111', 'dddddddd-dddd-dddd-dddd-dddddddddddd', '6a6a6a6a-6a6a-6a6a-6a6a-6a6a6a6a6a6a', NULL, 4, 'Смена на экологической акции: сбор и сортировка вторсырья.', 'approved', 'cccccccc-cccc-cccc-cccc-cccccccccccc', now() - interval '17 days', now() - interval '18 days', now() - interval '17 days'),
	('77777777-8888-7777-8888-777777777777', '11111111-1111-1111-1111-111111111111', 'dddddddd-dddd-dddd-dddd-dddddddddddd', '6c6c6c6c-6c6c-6c6c-6c6c-6c6c6c6c6c6c', NULL, 1.5, 'Участие в онлайн-инструктаже волонтеров.', 'approved', 'cccccccc-cccc-cccc-cccc-cccccccccccc', now() - interval '4 days', now() - interval '5 days', now() - interval '4 days'),
	('78787878-9999-7878-9999-787878787878', '11111111-1111-1111-1111-111111111111', 'dddddddd-dddd-dddd-dddd-dddddddddddd', NULL, '73737373-7373-7373-7373-737373737373', 2.25, 'Проверил и сверил список получателей наборов.', 'pending', NULL, NULL, now() - interval '90 minutes', now() - interval '90 minutes'),
	('79797979-7979-7979-7979-797979797979', '11111111-1111-1111-1111-111111111111', 'dddddddd-dddd-dddd-dddd-dddddddddddd', NULL, '74747474-7474-7474-7474-747474747474', 3, 'Подготовил коробки и маркировку для зоны выдачи.', 'pending', NULL, NULL, now() - interval '35 minutes', now() - interval '35 minutes'),
	('7a7a7a7a-7a7a-7a7a-7a7a-7a7a7a7a7a7a', '11111111-1111-1111-1111-111111111111', 'dddddddd-dddd-dddd-dddd-dddddddddddd', '6b6b6b6b-6b6b-6b6b-6b6b-6b6b6b6b6b6b', NULL, 1, 'Пробная запись времени для будущей смены.', 'pending', NULL, NULL, now() - interval '20 minutes', now() - interval '20 minutes'),
	('7b7b7b7b-7b7b-7b7b-7b7b-7b7b7b7b7b7b', '11111111-1111-1111-1111-111111111111', 'cccccccc-cccc-cccc-cccc-cccccccccccc', NULL, '73737373-7373-7373-7373-737373737373', 1.75, 'Координация списка и распределение задач по волонтерам.', 'approved', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', now() - interval '1 day', now() - interval '2 days', now() - interval '1 day')
ON CONFLICT (id) DO UPDATE SET
	hours = EXCLUDED.hours,
	description = EXCLUDED.description,
	status = EXCLUDED.status,
	reviewed_by = EXCLUDED.reviewed_by,
	reviewed_at = EXCLUDED.reviewed_at,
	updated_at = EXCLUDED.updated_at;

INSERT INTO shop_products (id, organization_id, name, description, price, stock, image_url, is_active)
VALUES
	('7c7c7c7c-7c7c-7c7c-7c7c-7c7c7c7c7c7c', '11111111-1111-1111-1111-111111111111', 'Фирменный блокнот', 'Блокнот для смен, чек-листов и заметок координатора.', 90, 35, 'https://images.unsplash.com/photo-1517842645767-c639042777db?auto=format&fit=crop&w=800&q=80', true),
	('7d7d7d7d-7d7d-7d7d-7d7d-7d7d7d7d7d7d', '11111111-1111-1111-1111-111111111111', 'Пауэрбанк для смены', 'Компактный внешний аккумулятор для длинных выездных мероприятий.', 420, 6, 'https://images.unsplash.com/photo-1609091839311-d5365f9ff1c5?auto=format&fit=crop&w=800&q=80', true),
	('7e7e7e7e-7e7e-7e7e-7e7e-7e7e7e7e7e7e', '11111111-1111-1111-1111-111111111111', 'Набор стикеров Пульс', 'Стикеры для ноутбука, бейджа и волонтерского дневника.', 45, 70, 'https://images.unsplash.com/photo-1513364776144-60967b0f800f?auto=format&fit=crop&w=800&q=80', true)
ON CONFLICT (id) DO UPDATE SET
	name = EXCLUDED.name,
	description = EXCLUDED.description,
	price = EXCLUDED.price,
	stock = EXCLUDED.stock,
	image_url = EXCLUDED.image_url,
	is_active = EXCLUDED.is_active,
	updated_at = now();

INSERT INTO shop_orders (id, user_id, organization_id, status, total, comment, processed_by, processed_at, created_at, updated_at)
VALUES
	('7f7f7f7f-7f7f-7f7f-7f7f-7f7f7f7f7f7f', 'dddddddd-dddd-dddd-dddd-dddddddddddd', '11111111-1111-1111-1111-111111111111', 'completed', 90, 'Блокнот выдан после инструктажа.', 'cccccccc-cccc-cccc-cccc-cccccccccccc', now() - interval '3 days', now() - interval '4 days', now() - interval '3 days'),
	('80808080-8080-8080-8080-808080808080', 'dddddddd-dddd-dddd-dddd-dddddddddddd', '11111111-1111-1111-1111-111111111111', 'pending', 45, 'Заберу на следующей смене.', NULL, NULL, now() - interval '25 minutes', now() - interval '25 minutes')
ON CONFLICT (id) DO UPDATE SET
	status = EXCLUDED.status,
	total = EXCLUDED.total,
	comment = EXCLUDED.comment,
	processed_by = EXCLUDED.processed_by,
	processed_at = EXCLUDED.processed_at,
	updated_at = EXCLUDED.updated_at;

DELETE FROM shop_order_items
WHERE order_id IN ('7f7f7f7f-7f7f-7f7f-7f7f-7f7f7f7f7f7f', '80808080-8080-8080-8080-808080808080');

INSERT INTO shop_order_items (order_id, product_id, product_name, price, quantity)
VALUES
	('7f7f7f7f-7f7f-7f7f-7f7f-7f7f7f7f7f7f', '7c7c7c7c-7c7c-7c7c-7c7c-7c7c7c7c7c7c', 'Фирменный блокнот', 90, 1),
	('80808080-8080-8080-8080-808080808080', '7e7e7e7e-7e7e-7e7e-7e7e-7e7e7e7e7e7e', 'Набор стикеров Пульс', 45, 1);

INSERT INTO coin_transactions (user_id, amount, type, description, source_type, source_id)
VALUES
	('dddddddd-dddd-dddd-dddd-dddddddddddd', -90, 'purchase', 'Демо-заказ: фирменный блокнот', 'shop_order', '7f7f7f7f-7f7f-7f7f-7f7f-7f7f7f7f7f7f'),
	('dddddddd-dddd-dddd-dddd-dddddddddddd', -45, 'purchase', 'Демо-заказ: набор стикеров', 'shop_order', '80808080-8080-8080-8080-808080808080')
ON CONFLICT (user_id, type, source_type, source_id) DO UPDATE SET
	amount = EXCLUDED.amount,
	description = EXCLUDED.description;

INSERT INTO certificates (id, user_id, organization_id, type, title, description, total_hours, verify_code, issued_by, issued_at)
VALUES
	('81818181-8181-8181-8181-818181818181', 'dddddddd-dddd-dddd-dddd-dddddddddddd', '11111111-1111-1111-1111-111111111111', 'hours', 'Справка о волонтерских часах', 'Подтверждает суммарный вклад волонтера по утвержденным записям времени.', 0, 'PULS-HOURS-2026', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', now() - interval '2 days'),
	('82828282-8282-8282-8282-828282828282', 'dddddddd-dddd-dddd-dddd-dddddddddddd', '11111111-1111-1111-1111-111111111111', 'participation', 'Сертификат об участии', 'Подтверждает участие волонтера в мероприятиях организации.', 0, 'PULS-PART-2026', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', now() - interval '1 day')
ON CONFLICT (verify_code) DO UPDATE SET
	title = EXCLUDED.title,
	description = EXCLUDED.description,
	type = EXCLUDED.type,
	issued_by = EXCLUDED.issued_by,
	issued_at = EXCLUDED.issued_at;

SELECT recalculate_user_gamification(id)
FROM users
WHERE email IN ('admin@puls.test', 'admin1@puls.test', 'admin2@puls.test', 'admin3@puls.test');

UPDATE volunteer_profiles vp
SET coin_balance = COALESCE(ledger.balance, 0),
	updated_at = now()
FROM (
	SELECT user_id, SUM(amount)::int AS balance
	FROM coin_transactions
	GROUP BY user_id
) ledger
WHERE ledger.user_id = vp.user_id
	AND vp.user_id IN (
		'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
		'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
		'cccccccc-cccc-cccc-cccc-cccccccccccc',
		'dddddddd-dddd-dddd-dddd-dddddddddddd'
	);
