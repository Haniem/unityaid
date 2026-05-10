-- Генерируем bcrypt-хэши средствами PostgreSQL, чтобы демо-пароль точно совпадал.
-- Пароль для всех демо-пользователей: password
UPDATE users
SET
	password_hash = crypt('password', gen_salt('bf', 10)),
	is_active = true,
	is_email_verified = true,
	updated_at = now()
WHERE email IN (
	'admin@unityaid.test',
	'org.admin@dobrye-ruki.test',
	'coord@dobrye-ruki.test',
	'volunteer1@test.local',
	'volunteer2@test.local',
	'org.admin@ecopulse.test',
	'coord@ecopulse.test',
	'volunteer3@test.local'
);
