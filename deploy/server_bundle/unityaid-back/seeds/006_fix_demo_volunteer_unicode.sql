UPDATE users
SET
	first_name = 'Илья',
	last_name = 'Морозов',
	updated_at = now()
WHERE email = 'admin3@puls.test'
;

UPDATE volunteer_profiles
SET
	city = 'Екатеринбург',
	bio = 'Готов помогать на городских событиях и социальных акциях.',
	updated_at = now()
WHERE user_id = 'dddddddd-dddd-dddd-dddd-dddddddddddd'
;
