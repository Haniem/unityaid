CREATE OR REPLACE FUNCTION ensure_display_id(target_table regclass)
RETURNS void AS $$
DECLARE
	sequence_name text := replace(target_table::text, '.', '_') || '_display_id_seq';
	index_name text := replace(target_table::text, '.', '_') || '_display_id_key';
BEGIN
	IF NOT EXISTS (
		SELECT 1
		FROM pg_attribute
		WHERE attrelid = target_table
			AND attname = 'display_id'
			AND NOT attisdropped
	) THEN
		EXECUTE format('ALTER TABLE %s ADD COLUMN display_id BIGINT', target_table);
	END IF;

	EXECUTE format('CREATE SEQUENCE IF NOT EXISTS %I', sequence_name);
	EXECUTE format('ALTER SEQUENCE %I OWNED BY %s.display_id', sequence_name, target_table);
	EXECUTE format('ALTER TABLE %s ALTER COLUMN display_id SET DEFAULT nextval(%L::regclass)', target_table, sequence_name);

	EXECUTE format(
		'WITH numbered AS (
			SELECT ctid, row_number() OVER (ORDER BY ctid) AS rn
			FROM %s
			WHERE display_id IS NULL
		)
		UPDATE %s AS t
		SET display_id = numbered.rn
		FROM numbered
		WHERE t.ctid = numbered.ctid',
		target_table,
		target_table
	);

	EXECUTE format(
		'SELECT setval(%L, GREATEST(COALESCE((SELECT max(display_id) FROM %s), 0) + 1, 1), false)',
		sequence_name,
		target_table
	);

	EXECUTE format('ALTER TABLE %s ALTER COLUMN display_id SET NOT NULL', target_table);
	EXECUTE format('CREATE UNIQUE INDEX IF NOT EXISTS %I ON %s (display_id)', index_name, target_table);
END;
$$ LANGUAGE plpgsql;

SELECT ensure_display_id('organizations'::regclass);
SELECT ensure_display_id('users'::regclass);
SELECT ensure_display_id('organization_members'::regclass);
SELECT ensure_display_id('volunteer_profiles'::regclass);
SELECT ensure_display_id('skills'::regclass);
SELECT ensure_display_id('volunteer_skills'::regclass);
SELECT ensure_display_id('events'::regclass);
SELECT ensure_display_id('tasks'::regclass);
SELECT ensure_display_id('task_assignments'::regclass);
SELECT ensure_display_id('news'::regclass);
SELECT ensure_display_id('refresh_sessions'::regclass);
SELECT ensure_display_id('revoked_access_tokens'::regclass);
SELECT ensure_display_id('email_verification_tokens'::regclass);
SELECT ensure_display_id('password_reset_tokens'::regclass);
SELECT ensure_display_id('event_applications'::regclass);
SELECT ensure_display_id('event_attendance'::regclass);
SELECT ensure_display_id('event_shifts'::regclass);
SELECT ensure_display_id('event_feedback'::regclass);
SELECT ensure_display_id('task_comments'::regclass);
SELECT ensure_display_id('task_attachments'::regclass);
SELECT ensure_display_id('task_status_history'::regclass);
SELECT ensure_display_id('task_time_entries'::regclass);
SELECT ensure_display_id('news_categories'::regclass);
SELECT ensure_display_id('system_roles'::regclass);
SELECT ensure_display_id('user_system_roles'::regclass);
SELECT ensure_display_id('time_entries'::regclass);
SELECT ensure_display_id('achievements'::regclass);
SELECT ensure_display_id('achievement_rules'::regclass);
SELECT ensure_display_id('volunteer_achievements'::regclass);
SELECT ensure_display_id('points_transactions'::regclass);
SELECT ensure_display_id('audit_log'::regclass);
SELECT ensure_display_id('notifications'::regclass);
SELECT ensure_display_id('knowledge_categories'::regclass);
SELECT ensure_display_id('knowledge_articles'::regclass);
SELECT ensure_display_id('certificates'::regclass);
SELECT ensure_display_id('field_qr_tokens'::regclass);
SELECT ensure_display_id('field_checkins'::regclass);
SELECT ensure_display_id('tenant_settings'::regclass);
SELECT ensure_display_id('user_invitations'::regclass);
SELECT ensure_display_id('event_templates'::regclass);

DROP FUNCTION ensure_display_id(regclass);
