-- Генерируем bcrypt-хэши средствами PostgreSQL, чтобы демо-пароль точно совпадал.
-- Пароль для всех демо-пользователей: password
UPDATE users
SET
	password_hash = crypt('password', gen_salt('bf', 10)),
	is_active = true,
	is_email_verified = true,
	updated_at = now()
WHERE email IN (
	'admin@puls.test',
	'admin1@puls.test',
	'admin2@puls.test',
	'admin3@puls.test'
);
