CREATE TABLE IF NOT EXISTS profile_field_groups (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
	code TEXT NOT NULL,
	name TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	sort_order INTEGER NOT NULL DEFAULT 0,
	is_system BOOLEAN NOT NULL DEFAULT false,
	is_active BOOLEAN NOT NULL DEFAULT true,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	UNIQUE (organization_id, code)
);

CREATE TABLE IF NOT EXISTS profile_field_definitions (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
	group_id UUID NOT NULL REFERENCES profile_field_groups(id) ON DELETE CASCADE,
	code TEXT NOT NULL,
	name TEXT NOT NULL,
	field_type TEXT NOT NULL CHECK (field_type IN ('text', 'textarea', 'number', 'date', 'datetime', 'tel', 'email', 'url', 'select', 'multiselect', 'checkbox', 'file')),
	required BOOLEAN NOT NULL DEFAULT false,
	is_system BOOLEAN NOT NULL DEFAULT false,
	is_active BOOLEAN NOT NULL DEFAULT true,
	editable_by_user BOOLEAN NOT NULL DEFAULT true,
	sort_order INTEGER NOT NULL DEFAULT 0,
	placeholder TEXT NOT NULL DEFAULT '',
	help TEXT NOT NULL DEFAULT '',
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	UNIQUE (organization_id, code)
);

CREATE TABLE IF NOT EXISTS profile_field_options (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	field_id UUID NOT NULL REFERENCES profile_field_definitions(id) ON DELETE CASCADE,
	value TEXT NOT NULL,
	label TEXT NOT NULL,
	sort_order INTEGER NOT NULL DEFAULT 0,
	is_active BOOLEAN NOT NULL DEFAULT true,
	UNIQUE (field_id, value)
);

CREATE TABLE IF NOT EXISTS profile_field_values (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
	user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	field_id UUID NOT NULL REFERENCES profile_field_definitions(id) ON DELETE CASCADE,
	value JSONB NOT NULL DEFAULT 'null'::jsonb,
	updated_by UUID REFERENCES users(id) ON DELETE SET NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	UNIQUE (organization_id, user_id, field_id)
);

CREATE INDEX IF NOT EXISTS idx_profile_groups_organization ON profile_field_groups(organization_id, sort_order);
CREATE INDEX IF NOT EXISTS idx_profile_fields_group ON profile_field_definitions(group_id, sort_order);
CREATE INDEX IF NOT EXISTS idx_profile_values_user ON profile_field_values(organization_id, user_id);

INSERT INTO profile_field_groups (organization_id, code, name, sort_order, is_system)
SELECT id, seed.code, seed.name, seed.sort_order, true
FROM organizations
CROSS JOIN (VALUES
	('personal', 'Личная информация', 10),
	('contacts', 'Контакты', 20),
	('work', 'Работа', 30),
	('results', 'Результаты', 40)
) AS seed(code, name, sort_order)
ON CONFLICT (organization_id, code) DO NOTHING;

INSERT INTO profile_field_definitions (organization_id, group_id, code, name, field_type, sort_order, is_system, editable_by_user)
SELECT g.organization_id, g.id, seed.code, seed.name, seed.field_type, seed.sort_order, true, seed.editable_by_user
FROM profile_field_groups g
JOIN (VALUES
	('personal', 'firstName', 'Имя', 'text', 10, true),
	('personal', 'lastName', 'Фамилия', 'text', 20, true),
	('personal', 'patronymic', 'Отчество', 'text', 30, true),
	('contacts', 'email', 'Электронная почта', 'email', 10, false),
	('contacts', 'phone', 'Телефон', 'tel', 20, true),
	('work', 'city', 'Город', 'text', 10, true),
	('work', 'bio', 'Опыт и описание', 'textarea', 20, true),
	('work', 'interests', 'Интересы', 'textarea', 30, true),
	('results', 'totalHours', 'Подтвержденные часы', 'number', 10, false),
	('results', 'points', 'Баллы', 'number', 20, false)
) AS seed(group_code, code, name, field_type, sort_order, editable_by_user)
	ON seed.group_code = g.code
WHERE g.is_system = true
ON CONFLICT (organization_id, code) DO NOTHING;
