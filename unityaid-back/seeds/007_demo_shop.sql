INSERT INTO shop_products (id, organization_id, name, description, price, stock, image_url, is_active)
VALUES
	('31313131-3131-3131-3131-313131313131', '11111111-1111-1111-1111-111111111111', 'Худи Пульс Добра', 'Теплое худи для волонтерских выездов и командных мероприятий.', 320, 12, 'https://images.unsplash.com/photo-1556821840-3a63f95609a7?auto=format&fit=crop&w=800&q=80', true),
	('32323232-3232-3232-3232-323232323232', '11111111-1111-1111-1111-111111111111', 'Термокружка', 'Стальная кружка с логотипом для длинных смен и полевых штабов.', 180, 24, 'https://images.unsplash.com/photo-1577937927133-66ef06acdf18?auto=format&fit=crop&w=800&q=80', true),
	('34343434-3434-3434-3434-343434343434', '11111111-1111-1111-1111-111111111111', 'Набор значков', 'Набор из четырех значков за участие в городских акциях.', 75, 40, 'https://images.unsplash.com/photo-1523293915678-d126868e96f1?auto=format&fit=crop&w=800&q=80', true),
	('35353535-3535-3535-3535-353535353535', '11111111-1111-1111-1111-111111111111', 'Сертификат на обучение', 'Оплата участия во внутреннем тренинге для координаторов.', 260, 8, 'https://images.unsplash.com/photo-1516321318423-f06f85e504b3?auto=format&fit=crop&w=800&q=80', true),
	('36363636-3636-3636-3636-363636363636', '11111111-1111-1111-1111-111111111111', 'Эко-шоппер', 'Плотная сумка для сборов, выдач и личных вещей на мероприятиях.', 120, 30, 'https://images.unsplash.com/photo-1597484661643-2f5fef640dd1?auto=format&fit=crop&w=800&q=80', true)
ON CONFLICT (id) DO UPDATE SET
	name = EXCLUDED.name,
	description = EXCLUDED.description,
	price = EXCLUDED.price,
	stock = EXCLUDED.stock,
	image_url = EXCLUDED.image_url,
	is_active = EXCLUDED.is_active,
	updated_at = now();

INSERT INTO coin_transactions (user_id, amount, type, description, source_type, source_id)
VALUES
	('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 900, 'demo_bonus', 'Стартовый баланс главного администратора', 'demo_seed', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'),
	('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 700, 'demo_bonus', 'Стартовый баланс администратора организации', 'demo_seed', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb'),
	('cccccccc-cccc-cccc-cccc-cccccccccccc', 650, 'demo_bonus', 'Стартовый баланс менеджера организации', 'demo_seed', 'cccccccc-cccc-cccc-cccc-cccccccccccc'),
	('dddddddd-dddd-dddd-dddd-dddddddddddd', 520, 'demo_bonus', 'Стартовый баланс волонтера', 'demo_seed', 'dddddddd-dddd-dddd-dddd-dddddddddddd')
ON CONFLICT (user_id, type, source_type, source_id) DO UPDATE SET
	amount = EXCLUDED.amount,
	description = EXCLUDED.description;

INSERT INTO shop_orders (id, user_id, organization_id, status, total, comment, processed_by, processed_at, created_at, updated_at)
VALUES
	('37373737-3737-3737-3737-373737373737', 'dddddddd-dddd-dddd-dddd-dddddddddddd', '11111111-1111-1111-1111-111111111111', 'pending', 255, 'Заберу после инструктажа.', NULL, NULL, now() - interval '2 hours', now() - interval '2 hours'),
	('38383838-3838-3838-3838-383838383838', 'cccccccc-cccc-cccc-cccc-cccccccccccc', '11111111-1111-1111-1111-111111111111', 'processing', 260, 'Для подготовки координаторов.', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', now() - interval '1 hour', now() - interval '1 day', now() - interval '1 hour'),
	('39393939-3939-3939-3939-393939393939', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '11111111-1111-1111-1111-111111111111', 'completed', 180, 'Получено в офисе.', 'cccccccc-cccc-cccc-cccc-cccccccccccc', now() - interval '3 days', now() - interval '4 days', now() - interval '3 days')
ON CONFLICT (id) DO UPDATE SET
	status = EXCLUDED.status,
	total = EXCLUDED.total,
	comment = EXCLUDED.comment,
	processed_by = EXCLUDED.processed_by,
	processed_at = EXCLUDED.processed_at,
	updated_at = EXCLUDED.updated_at;

DELETE FROM shop_order_items
WHERE order_id IN (
	'37373737-3737-3737-3737-373737373737',
	'38383838-3838-3838-3838-383838383838',
	'39393939-3939-3939-3939-393939393939'
);

INSERT INTO shop_order_items (order_id, product_id, product_name, price, quantity)
VALUES
	('37373737-3737-3737-3737-373737373737', '32323232-3232-3232-3232-323232323232', 'Термокружка', 180, 1),
	('37373737-3737-3737-3737-373737373737', '34343434-3434-3434-3434-343434343434', 'Набор значков', 75, 1),
	('38383838-3838-3838-3838-383838383838', '35353535-3535-3535-3535-353535353535', 'Сертификат на обучение', 260, 1),
	('39393939-3939-3939-3939-393939393939', '32323232-3232-3232-3232-323232323232', 'Термокружка', 180, 1);

INSERT INTO coin_transactions (user_id, amount, type, description, source_type, source_id)
VALUES
	('dddddddd-dddd-dddd-dddd-dddddddddddd', -255, 'purchase', 'Демо-заказ в корпоративном магазине', 'shop_order', '37373737-3737-3737-3737-373737373737'),
	('cccccccc-cccc-cccc-cccc-cccccccccccc', -260, 'purchase', 'Демо-заказ в корпоративном магазине', 'shop_order', '38383838-3838-3838-3838-383838383838'),
	('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', -180, 'purchase', 'Демо-заказ в корпоративном магазине', 'shop_order', '39393939-3939-3939-3939-393939393939')
ON CONFLICT (user_id, type, source_type, source_id) DO UPDATE SET
	amount = EXCLUDED.amount,
	description = EXCLUDED.description;

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
