-- Геймификация и компетенции
-- Таблиц: 8

CREATE TABLE achievements (
  id uuid NOT NULL,
  code text NOT NULL,
  name text NOT NULL,
  icon text NOT NULL,
  points_reward integer NOT NULL,
  PRIMARY KEY (id)
);

CREATE TABLE skills (
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

CREATE TABLE achievement_rules (
  id uuid NOT NULL,
  achievement_id uuid NOT NULL REFERENCES achievements(id),
  code text NOT NULL,
  trigger_type text NOT NULL,
  threshold numeric,
  points integer NOT NULL,
  PRIMARY KEY (id)
);

CREATE TABLE volunteer_profiles (
  id uuid NOT NULL,
  user_id uuid NOT NULL REFERENCES users(id),
  city text,
  points integer NOT NULL,
  level integer NOT NULL,
  status text NOT NULL,
  PRIMARY KEY (id)
);

CREATE TABLE volunteer_achievements (
  id uuid NOT NULL,
  user_id uuid NOT NULL REFERENCES users(id),
  achievement_id uuid NOT NULL REFERENCES achievements(id),
  earned_at timestamptz NOT NULL,
  PRIMARY KEY (id)
);

CREATE TABLE points_transactions (
  id uuid NOT NULL,
  user_id uuid NOT NULL REFERENCES users(id),
  achievement_id uuid REFERENCES achievements(id),
  source_type text NOT NULL,
  source_id uuid,
  points integer NOT NULL,
  reason text NOT NULL,
  PRIMARY KEY (id)
);

CREATE TABLE volunteer_skills (
  volunteer_profile_id uuid NOT NULL REFERENCES volunteer_profiles(id),
  skill_id uuid NOT NULL REFERENCES skills(id),
  PRIMARY KEY (volunteer_profile_id, skill_id)
);
