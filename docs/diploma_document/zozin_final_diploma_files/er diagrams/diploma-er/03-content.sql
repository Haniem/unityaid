-- Контент и уведомления
-- Таблиц: 7

CREATE TABLE knowledge_categories (
  id uuid NOT NULL,
  name text NOT NULL,
  PRIMARY KEY (id)
);

CREATE TABLE news_categories (
  id uuid NOT NULL,
  name text NOT NULL,
  PRIMARY KEY (id)
);

CREATE TABLE organizations (
  id uuid NOT NULL,
  name text NOT NULL,
  PRIMARY KEY (id)
);

CREATE TABLE users (
  id uuid NOT NULL,
  email text NOT NULL,
  first_name text NOT NULL,
  last_name text NOT NULL,
  PRIMARY KEY (id)
);

CREATE TABLE notifications (
  id uuid NOT NULL,
  user_id uuid NOT NULL REFERENCES users(id),
  type text NOT NULL,
  title text NOT NULL,
  body text NOT NULL,
  is_read boolean NOT NULL,
  PRIMARY KEY (id)
);

CREATE TABLE knowledge_articles (
  id uuid NOT NULL,
  category_id uuid REFERENCES knowledge_categories(id),
  title text NOT NULL,
  status text NOT NULL,
  author_id uuid REFERENCES users(id),
  PRIMARY KEY (id)
);

CREATE TABLE news (
  id uuid NOT NULL,
  organization_id uuid REFERENCES organizations(id),
  title text NOT NULL,
  status text NOT NULL,
  author_id uuid REFERENCES users(id),
  category_id uuid REFERENCES news_categories(id),
  PRIMARY KEY (id)
);
