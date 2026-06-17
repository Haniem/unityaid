CREATE TABLE IF NOT EXISTS knowledge_categories (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	name TEXT NOT NULL UNIQUE,
	slug TEXT NOT NULL UNIQUE,
	description TEXT NOT NULL DEFAULT '',
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS knowledge_articles (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	category_id UUID REFERENCES knowledge_categories(id) ON DELETE SET NULL,
	title TEXT NOT NULL,
	slug TEXT NOT NULL UNIQUE,
	summary TEXT NOT NULL DEFAULT '',
	content_html TEXT NOT NULL DEFAULT '',
	status TEXT NOT NULL DEFAULT 'draft',
	author_id UUID REFERENCES users(id) ON DELETE SET NULL,
	published_at TIMESTAMPTZ,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_knowledge_articles_category_id ON knowledge_articles(category_id);
CREATE INDEX IF NOT EXISTS idx_knowledge_articles_status ON knowledge_articles(status);
CREATE INDEX IF NOT EXISTS idx_knowledge_articles_title ON knowledge_articles USING gin(to_tsvector('simple', title || ' ' || summary || ' ' || content_html));

INSERT INTO knowledge_categories (name, slug, description)
VALUES
	('Старт волонтера', 'start-volunteer', 'Материалы для знакомства с платформой и первыми шагами.'),
	('Мероприятия', 'events', 'Инструкции по участию в мероприятиях.'),
	('Задачи', 'tasks', 'Справка по задачам и отчетности.')
ON CONFLICT (slug) DO NOTHING;
