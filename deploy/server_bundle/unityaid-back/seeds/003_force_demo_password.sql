-- Пароль для всех демо-пользователей: password
UPDATE users
SET
	password_hash = '$2a$10$uISbNqAL/.aA87KggeAr6OohJh9J.uY2i7O8k7qJ5WqsPBFO9/ARa',
	is_active = true,
	is_email_verified = true,
	updated_at = now()
WHERE email IN (
	'admin@puls.test',
	'admin1@puls.test',
	'admin2@puls.test',
	'admin3@puls.test'
);
