--
-- PostgreSQL database dump
--

\restrict 5Uaf3Y0gCcDOfpwBX5Hjf4hrdLAk40jx15k9rfd9mrqtWSIzviQfDI9n9pT8v1b

-- Dumped from database version 17.9
-- Dumped by pg_dump version 17.9

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: pgcrypto; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public;


--
-- Name: EXTENSION pgcrypto; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION pgcrypto IS 'cryptographic functions';


--
-- Name: app_role; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.app_role AS ENUM (
    'super_admin',
    'org_admin',
    'coordinator',
    'volunteer'
);


--
-- Name: event_application_status; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.event_application_status AS ENUM (
    'pending',
    'approved',
    'waitlisted',
    'rejected',
    'cancelled'
);


--
-- Name: event_format; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.event_format AS ENUM (
    'online',
    'offline',
    'hybrid'
);


--
-- Name: event_status; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.event_status AS ENUM (
    'draft',
    'published',
    'completed',
    'cancelled'
);


--
-- Name: member_status; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.member_status AS ENUM (
    'active',
    'inactive',
    'blocked'
);


--
-- Name: task_status; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.task_status AS ENUM (
    'created',
    'assigned',
    'in_progress',
    'review',
    'completed',
    'cancelled'
);


--
-- Name: volunteer_status; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.volunteer_status AS ENUM (
    'new',
    'active',
    'unavailable',
    'archived'
);


--
-- Name: create_achievement_notification(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.create_achievement_notification() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
	achievement_name TEXT;
BEGIN
	SELECT name INTO achievement_name FROM achievements WHERE id = NEW.achievement_id;

	INSERT INTO notifications (user_id, type, title, body, link, entity_type, entity_id)
	VALUES (
		NEW.user_id,
		'achievement_earned',
		'Достижение получено',
		COALESCE(achievement_name, ''),
		'/achievements',
		'achievement',
		NEW.achievement_id::text
	)
	ON CONFLICT (user_id, type, entity_type, entity_id)
	DO UPDATE SET
		title = EXCLUDED.title,
		body = EXCLUDED.body,
		link = EXCLUDED.link,
		is_read = false,
		read_at = NULL,
		created_at = now();

	RETURN NEW;
END;
$$;


--
-- Name: parse_period_from(text); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.parse_period_from(value text) RETURNS timestamp with time zone
    LANGUAGE sql STABLE
    AS $$
	SELECT COALESCE(NULLIF(value, '')::timestamptz, now() - interval '30 days')
$$;


--
-- Name: parse_period_to(text); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.parse_period_to(value text) RETURNS timestamp with time zone
    LANGUAGE sql STABLE
    AS $$
	SELECT COALESCE(NULLIF(value, '')::timestamptz, now())
$$;


--
-- Name: recalculate_user_gamification(uuid); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.recalculate_user_gamification(target_user_id uuid) RETURNS void
    LANGUAGE plpgsql
    AS $$
DECLARE
	total_points INTEGER;
BEGIN
	INSERT INTO volunteer_profiles (user_id)
	VALUES (target_user_id)
	ON CONFLICT (user_id) DO NOTHING;

	WITH eligible AS (
		SELECT a.id, a.points_reward, a.name
		FROM achievements a
		JOIN volunteer_profiles vp ON vp.user_id = target_user_id
		JOIN users u ON u.id = target_user_id
		WHERE
			(a.code = 'registration')
			OR (
				a.code = 'profile_completed'
				AND (
					COALESCE(vp.bio, '') <> ''
					OR COALESCE(vp.city, '') <> ''
					OR COALESCE(vp.phone, '') <> ''
					OR COALESCE(vp.interests, '') <> ''
				)
			)
			OR (
				a.code = 'first_participation'
				AND EXISTS (
					SELECT 1 FROM event_attendance ea WHERE ea.user_id = target_user_id
				)
			)
			OR (
				a.code = 'task_completed'
				AND EXISTS (
					SELECT 1
					FROM task_assignments ta
					JOIN tasks t ON t.id = ta.task_id
					WHERE ta.user_id = target_user_id AND t.status = 'completed'
				)
			)
			OR (a.code = 'hours_10' AND vp.total_hours >= 10)
			OR (a.code = 'hours_25' AND vp.total_hours >= 25)
			OR (a.code = 'hours_50' AND vp.total_hours >= 50)
			OR (a.code = 'hours_100' AND vp.total_hours >= 100)
	),
	new_awards AS (
		INSERT INTO volunteer_achievements (user_id, achievement_id)
		SELECT target_user_id, id FROM eligible
		ON CONFLICT (user_id, achievement_id) DO NOTHING
		RETURNING achievement_id
	)
	INSERT INTO points_transactions (user_id, achievement_id, points, reason)
	SELECT target_user_id, a.id, a.points_reward, a.name
	FROM achievements a
	JOIN new_awards na ON na.achievement_id = a.id
	WHERE a.points_reward <> 0
	ON CONFLICT (user_id, achievement_id, source_type) DO NOTHING;

	SELECT COALESCE(SUM(a.points_reward), 0)
	INTO total_points
	FROM volunteer_achievements va
	JOIN achievements a ON a.id = va.achievement_id
	WHERE va.user_id = target_user_id;

	UPDATE volunteer_profiles
	SET points = total_points,
		level = GREATEST(1, FLOOR(total_points / 100.0)::INTEGER + 1),
		updated_at = now()
	WHERE user_id = target_user_id;
END;
$$;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: achievement_rules; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.achievement_rules (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    achievement_id uuid NOT NULL,
    code text NOT NULL,
    trigger_type text NOT NULL,
    threshold numeric(10,2),
    points integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    display_id bigint NOT NULL
);


--
-- Name: achievement_rules_display_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.achievement_rules_display_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: achievement_rules_display_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.achievement_rules_display_id_seq OWNED BY public.achievement_rules.display_id;


--
-- Name: achievements; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.achievements (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    code text NOT NULL,
    name text NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    icon text DEFAULT 'award'::text NOT NULL,
    points_reward integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    display_id bigint NOT NULL
);


--
-- Name: achievements_display_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.achievements_display_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: achievements_display_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.achievements_display_id_seq OWNED BY public.achievements.display_id;


--
-- Name: audit_log; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.audit_log (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid,
    method text NOT NULL,
    action text NOT NULL,
    entity_type text NOT NULL,
    entity_id text,
    path text NOT NULL,
    status_code integer NOT NULL,
    ip_address text DEFAULT ''::text NOT NULL,
    user_agent text DEFAULT ''::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    display_id bigint NOT NULL
);


--
-- Name: audit_log_display_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.audit_log_display_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: audit_log_display_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.audit_log_display_id_seq OWNED BY public.audit_log.display_id;


--
-- Name: certificates; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.certificates (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    organization_id uuid,
    type text DEFAULT 'participation'::text NOT NULL,
    title text NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    total_hours numeric(10,2) DEFAULT 0 NOT NULL,
    verify_code text NOT NULL,
    file_path text DEFAULT ''::text NOT NULL,
    issued_by uuid,
    issued_at timestamp with time zone DEFAULT now() NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    display_id bigint NOT NULL
);


--
-- Name: certificates_display_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.certificates_display_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: certificates_display_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.certificates_display_id_seq OWNED BY public.certificates.display_id;


--
-- Name: email_verification_tokens; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.email_verification_tokens (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    token_hash text NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    used_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    display_id bigint NOT NULL
);


--
-- Name: email_verification_tokens_display_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.email_verification_tokens_display_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: email_verification_tokens_display_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.email_verification_tokens_display_id_seq OWNED BY public.email_verification_tokens.display_id;


--
-- Name: event_applications; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.event_applications (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    event_id uuid NOT NULL,
    user_id uuid NOT NULL,
    status public.event_application_status DEFAULT 'pending'::public.event_application_status NOT NULL,
    message text DEFAULT ''::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    rejection_reason text DEFAULT ''::text NOT NULL,
    display_id bigint NOT NULL
);


--
-- Name: event_applications_display_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.event_applications_display_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: event_applications_display_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.event_applications_display_id_seq OWNED BY public.event_applications.display_id;


--
-- Name: event_attendance; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.event_attendance (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    event_id uuid NOT NULL,
    user_id uuid NOT NULL,
    check_in_at timestamp with time zone,
    check_out_at timestamp with time zone,
    hours numeric(8,2) DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    display_id bigint NOT NULL
);


--
-- Name: event_attendance_display_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.event_attendance_display_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: event_attendance_display_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.event_attendance_display_id_seq OWNED BY public.event_attendance.display_id;


--
-- Name: event_feedback; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.event_feedback (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    event_id uuid NOT NULL,
    user_id uuid NOT NULL,
    rating integer NOT NULL,
    comment text DEFAULT ''::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    display_id bigint NOT NULL,
    CONSTRAINT event_feedback_rating_check CHECK (((rating >= 1) AND (rating <= 5)))
);


--
-- Name: event_feedback_display_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.event_feedback_display_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: event_feedback_display_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.event_feedback_display_id_seq OWNED BY public.event_feedback.display_id;


--
-- Name: event_shifts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.event_shifts (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    event_id uuid NOT NULL,
    title text NOT NULL,
    starts_at timestamp with time zone NOT NULL,
    ends_at timestamp with time zone NOT NULL,
    capacity integer,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    display_id bigint NOT NULL
);


--
-- Name: event_shifts_display_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.event_shifts_display_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: event_shifts_display_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.event_shifts_display_id_seq OWNED BY public.event_shifts.display_id;


--
-- Name: event_templates; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.event_templates (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    organization_id uuid NOT NULL,
    name text NOT NULL,
    title text NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    format public.event_format DEFAULT 'offline'::public.event_format NOT NULL,
    location text,
    max_participants integer,
    default_duration_minutes integer DEFAULT 120 NOT NULL,
    created_by uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    display_id bigint NOT NULL
);


--
-- Name: event_templates_display_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.event_templates_display_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: event_templates_display_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.event_templates_display_id_seq OWNED BY public.event_templates.display_id;


--
-- Name: events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.events (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    organization_id uuid NOT NULL,
    title text NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    format public.event_format NOT NULL,
    status public.event_status DEFAULT 'draft'::public.event_status NOT NULL,
    starts_at timestamp with time zone NOT NULL,
    ends_at timestamp with time zone NOT NULL,
    location text,
    max_participants integer,
    created_by uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    checkin_code text DEFAULT encode(public.gen_random_bytes(12), 'hex'::text) NOT NULL,
    display_id bigint NOT NULL
);


--
-- Name: events_display_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.events_display_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: events_display_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.events_display_id_seq OWNED BY public.events.display_id;


--
-- Name: field_checkins; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.field_checkins (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    event_id uuid NOT NULL,
    user_id uuid NOT NULL,
    qr_token_id uuid,
    checkin_at timestamp with time zone,
    checkout_at timestamp with time zone,
    status text DEFAULT 'checked_in'::text NOT NULL,
    source text DEFAULT 'qr'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    display_id bigint NOT NULL,
    CONSTRAINT field_checkins_status_check CHECK ((status = ANY (ARRAY['checked_in'::text, 'checked_out'::text])))
);


--
-- Name: field_checkins_display_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.field_checkins_display_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: field_checkins_display_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.field_checkins_display_id_seq OWNED BY public.field_checkins.display_id;


--
-- Name: field_qr_tokens; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.field_qr_tokens (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    event_id uuid NOT NULL,
    token text NOT NULL,
    mode text DEFAULT 'checkin'::text NOT NULL,
    expires_at timestamp with time zone,
    created_by uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    display_id bigint NOT NULL,
    CONSTRAINT field_qr_tokens_mode_check CHECK ((mode = ANY (ARRAY['checkin'::text, 'checkout'::text, 'attendance'::text])))
);


--
-- Name: field_qr_tokens_display_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.field_qr_tokens_display_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: field_qr_tokens_display_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.field_qr_tokens_display_id_seq OWNED BY public.field_qr_tokens.display_id;


--
-- Name: knowledge_articles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.knowledge_articles (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    category_id uuid,
    title text NOT NULL,
    slug text NOT NULL,
    summary text DEFAULT ''::text NOT NULL,
    content_html text DEFAULT ''::text NOT NULL,
    status text DEFAULT 'draft'::text NOT NULL,
    author_id uuid,
    published_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    display_id bigint NOT NULL
);


--
-- Name: knowledge_articles_display_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.knowledge_articles_display_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: knowledge_articles_display_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.knowledge_articles_display_id_seq OWNED BY public.knowledge_articles.display_id;


--
-- Name: knowledge_categories; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.knowledge_categories (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name text NOT NULL,
    slug text NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    display_id bigint NOT NULL
);


--
-- Name: knowledge_categories_display_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.knowledge_categories_display_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: knowledge_categories_display_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.knowledge_categories_display_id_seq OWNED BY public.knowledge_categories.display_id;


--
-- Name: news; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.news (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    organization_id uuid,
    title text NOT NULL,
    slug text NOT NULL,
    summary text DEFAULT ''::text NOT NULL,
    content_html text DEFAULT ''::text NOT NULL,
    cover_image_url text,
    status text DEFAULT 'published'::text NOT NULL,
    author_id uuid,
    published_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    category_id uuid,
    scheduled_at timestamp with time zone,
    display_id bigint NOT NULL
);


--
-- Name: news_categories; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.news_categories (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name text NOT NULL,
    slug text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    display_id bigint NOT NULL
);


--
-- Name: news_categories_display_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.news_categories_display_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: news_categories_display_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.news_categories_display_id_seq OWNED BY public.news_categories.display_id;


--
-- Name: news_display_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.news_display_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: news_display_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.news_display_id_seq OWNED BY public.news.display_id;


--
-- Name: notifications; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notifications (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    type text NOT NULL,
    title text NOT NULL,
    body text DEFAULT ''::text NOT NULL,
    link text DEFAULT ''::text NOT NULL,
    entity_type text DEFAULT ''::text NOT NULL,
    entity_id text DEFAULT ''::text NOT NULL,
    is_read boolean DEFAULT false NOT NULL,
    read_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    display_id bigint NOT NULL
);


--
-- Name: notifications_display_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.notifications_display_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: notifications_display_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.notifications_display_id_seq OWNED BY public.notifications.display_id;


--
-- Name: organization_members; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.organization_members (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    organization_id uuid NOT NULL,
    user_id uuid NOT NULL,
    role public.app_role NOT NULL,
    status public.member_status DEFAULT 'active'::public.member_status NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    display_id bigint NOT NULL
);


--
-- Name: organization_members_display_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.organization_members_display_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: organization_members_display_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.organization_members_display_id_seq OWNED BY public.organization_members.display_id;


--
-- Name: organizations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.organizations (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name text NOT NULL,
    slug text NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    contact_email text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    logo_url text,
    website_url text,
    phone text,
    address text,
    deleted_at timestamp with time zone,
    display_id bigint NOT NULL
);


--
-- Name: organizations_display_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.organizations_display_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: organizations_display_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.organizations_display_id_seq OWNED BY public.organizations.display_id;


--
-- Name: password_reset_tokens; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.password_reset_tokens (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    token_hash text NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    used_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    display_id bigint NOT NULL
);


--
-- Name: password_reset_tokens_display_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.password_reset_tokens_display_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: password_reset_tokens_display_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.password_reset_tokens_display_id_seq OWNED BY public.password_reset_tokens.display_id;


--
-- Name: points_transactions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.points_transactions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    achievement_id uuid,
    source_type text DEFAULT 'achievement'::text NOT NULL,
    source_id uuid,
    points integer NOT NULL,
    reason text DEFAULT ''::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    display_id bigint NOT NULL
);


--
-- Name: points_transactions_display_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.points_transactions_display_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: points_transactions_display_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.points_transactions_display_id_seq OWNED BY public.points_transactions.display_id;


--
-- Name: profile_field_definitions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.profile_field_definitions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    organization_id uuid NOT NULL,
    group_id uuid NOT NULL,
    code text NOT NULL,
    name text NOT NULL,
    field_type text NOT NULL,
    required boolean DEFAULT false NOT NULL,
    is_system boolean DEFAULT false NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    editable_by_user boolean DEFAULT true NOT NULL,
    sort_order integer DEFAULT 0 NOT NULL,
    placeholder text DEFAULT ''::text NOT NULL,
    help text DEFAULT ''::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT profile_field_definitions_field_type_check CHECK ((field_type = ANY (ARRAY['text'::text, 'textarea'::text, 'number'::text, 'date'::text, 'datetime'::text, 'tel'::text, 'email'::text, 'url'::text, 'select'::text, 'multiselect'::text, 'checkbox'::text, 'file'::text])))
);


--
-- Name: profile_field_groups; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.profile_field_groups (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    organization_id uuid NOT NULL,
    code text NOT NULL,
    name text NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    sort_order integer DEFAULT 0 NOT NULL,
    is_system boolean DEFAULT false NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: profile_field_options; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.profile_field_options (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    field_id uuid NOT NULL,
    value text NOT NULL,
    label text NOT NULL,
    sort_order integer DEFAULT 0 NOT NULL,
    is_active boolean DEFAULT true NOT NULL
);


--
-- Name: profile_field_values; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.profile_field_values (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    organization_id uuid NOT NULL,
    user_id uuid NOT NULL,
    field_id uuid NOT NULL,
    value jsonb DEFAULT 'null'::jsonb NOT NULL,
    updated_by uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: refresh_sessions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.refresh_sessions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    token_hash text NOT NULL,
    user_agent text DEFAULT ''::text NOT NULL,
    ip_address text DEFAULT ''::text NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    revoked_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    display_id bigint NOT NULL
);


--
-- Name: refresh_sessions_display_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.refresh_sessions_display_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: refresh_sessions_display_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.refresh_sessions_display_id_seq OWNED BY public.refresh_sessions.display_id;


--
-- Name: revoked_access_tokens; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.revoked_access_tokens (
    jti text NOT NULL,
    user_id uuid NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    revoked_at timestamp with time zone DEFAULT now() NOT NULL,
    display_id bigint NOT NULL
);


--
-- Name: revoked_access_tokens_display_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.revoked_access_tokens_display_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: revoked_access_tokens_display_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.revoked_access_tokens_display_id_seq OWNED BY public.revoked_access_tokens.display_id;


--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.schema_migrations (
    name text NOT NULL,
    applied_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: seed_migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.seed_migrations (
    name text NOT NULL,
    applied_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: skills; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.skills (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    display_id bigint NOT NULL
);


--
-- Name: skills_display_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.skills_display_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: skills_display_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.skills_display_id_seq OWNED BY public.skills.display_id;


--
-- Name: system_roles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.system_roles (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    code text NOT NULL,
    name text NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    display_id bigint NOT NULL
);


--
-- Name: system_roles_display_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.system_roles_display_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: system_roles_display_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.system_roles_display_id_seq OWNED BY public.system_roles.display_id;


--
-- Name: task_assignments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.task_assignments (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    task_id uuid NOT NULL,
    user_id uuid NOT NULL,
    assigned_at timestamp with time zone DEFAULT now() NOT NULL,
    role text DEFAULT 'assignee'::text NOT NULL,
    display_id bigint NOT NULL
);


--
-- Name: task_assignments_display_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.task_assignments_display_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: task_assignments_display_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.task_assignments_display_id_seq OWNED BY public.task_assignments.display_id;


--
-- Name: task_attachments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.task_attachments (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    task_id uuid NOT NULL,
    user_id uuid NOT NULL,
    file_name text NOT NULL,
    file_url text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    display_id bigint NOT NULL
);


--
-- Name: task_attachments_display_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.task_attachments_display_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: task_attachments_display_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.task_attachments_display_id_seq OWNED BY public.task_attachments.display_id;


--
-- Name: task_comments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.task_comments (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    task_id uuid NOT NULL,
    user_id uuid NOT NULL,
    content text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    display_id bigint NOT NULL
);


--
-- Name: task_comments_display_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.task_comments_display_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: task_comments_display_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.task_comments_display_id_seq OWNED BY public.task_comments.display_id;


--
-- Name: task_status_history; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.task_status_history (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    task_id uuid NOT NULL,
    from_status text,
    to_status text NOT NULL,
    changed_by uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    display_id bigint NOT NULL
);


--
-- Name: task_status_history_display_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.task_status_history_display_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: task_status_history_display_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.task_status_history_display_id_seq OWNED BY public.task_status_history.display_id;


--
-- Name: task_time_entries; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.task_time_entries (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    task_id uuid NOT NULL,
    user_id uuid NOT NULL,
    hours numeric(8,2) NOT NULL,
    note text DEFAULT ''::text NOT NULL,
    status text DEFAULT 'pending'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    display_id bigint NOT NULL
);


--
-- Name: task_time_entries_display_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.task_time_entries_display_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: task_time_entries_display_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.task_time_entries_display_id_seq OWNED BY public.task_time_entries.display_id;


--
-- Name: tasks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tasks (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    organization_id uuid NOT NULL,
    event_id uuid,
    title text NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    status public.task_status DEFAULT 'created'::public.task_status NOT NULL,
    priority text DEFAULT 'medium'::text NOT NULL,
    due_at timestamp with time zone,
    created_by uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    completion_confirmed_by uuid,
    completion_confirmed_at timestamp with time zone,
    display_id bigint NOT NULL
);


--
-- Name: tasks_display_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.tasks_display_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tasks_display_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.tasks_display_id_seq OWNED BY public.tasks.display_id;


--
-- Name: tenant_settings; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tenant_settings (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    display_name text DEFAULT 'UnityAid'::text NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    logo_url text,
    primary_color text DEFAULT '#2f9f72'::text NOT NULL,
    accent_color text DEFAULT '#22684e'::text NOT NULL,
    timezone text DEFAULT 'Asia/Yekaterinburg'::text NOT NULL,
    locale text DEFAULT 'ru'::text NOT NULL,
    contact_email text,
    contact_phone text,
    default_organization_id uuid,
    default_organization_name text DEFAULT ''::text NOT NULL,
    default_organization_slug text DEFAULT ''::text NOT NULL,
    onboarding_completed boolean DEFAULT false NOT NULL,
    pending_invites jsonb DEFAULT '[]'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    singleton_key boolean DEFAULT true NOT NULL,
    display_id bigint NOT NULL,
    CONSTRAINT tenant_settings_singleton CHECK (singleton_key)
);


--
-- Name: tenant_settings_display_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.tenant_settings_display_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tenant_settings_display_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.tenant_settings_display_id_seq OWNED BY public.tenant_settings.display_id;


--
-- Name: time_entries; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.time_entries (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    organization_id uuid NOT NULL,
    user_id uuid NOT NULL,
    event_id uuid,
    task_id uuid,
    hours numeric(8,2) NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    status text DEFAULT 'pending'::text NOT NULL,
    reviewed_by uuid,
    reviewed_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    display_id bigint NOT NULL,
    CONSTRAINT time_entries_check CHECK (((event_id IS NOT NULL) OR (task_id IS NOT NULL))),
    CONSTRAINT time_entries_hours_check CHECK ((hours > (0)::numeric)),
    CONSTRAINT time_entries_status_check CHECK ((status = ANY (ARRAY['pending'::text, 'approved'::text, 'rejected'::text])))
);


--
-- Name: time_entries_display_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.time_entries_display_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: time_entries_display_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.time_entries_display_id_seq OWNED BY public.time_entries.display_id;


--
-- Name: user_invitations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_invitations (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    email text NOT NULL,
    role text DEFAULT 'volunteer'::text NOT NULL,
    token text NOT NULL,
    status text DEFAULT 'pending'::text NOT NULL,
    invited_by uuid,
    accepted_by uuid,
    expires_at timestamp with time zone DEFAULT (now() + '14 days'::interval) NOT NULL,
    accepted_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    display_id bigint NOT NULL,
    CONSTRAINT user_invitations_status_check CHECK ((status = ANY (ARRAY['pending'::text, 'accepted'::text, 'revoked'::text, 'expired'::text])))
);


--
-- Name: user_invitations_display_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.user_invitations_display_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: user_invitations_display_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.user_invitations_display_id_seq OWNED BY public.user_invitations.display_id;


--
-- Name: user_system_roles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_system_roles (
    user_id uuid NOT NULL,
    role_id uuid NOT NULL,
    assigned_at timestamp with time zone DEFAULT now() NOT NULL,
    display_id bigint NOT NULL
);


--
-- Name: user_system_roles_display_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.user_system_roles_display_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: user_system_roles_display_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.user_system_roles_display_id_seq OWNED BY public.user_system_roles.display_id;


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    email text NOT NULL,
    password_hash text NOT NULL,
    first_name text NOT NULL,
    last_name text NOT NULL,
    patronymic text,
    avatar_url text,
    locale text DEFAULT 'ru'::text NOT NULL,
    is_email_verified boolean DEFAULT true NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    last_login_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    display_id bigint NOT NULL
);


--
-- Name: users_display_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.users_display_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: users_display_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.users_display_id_seq OWNED BY public.users.display_id;


--
-- Name: volunteer_achievements; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.volunteer_achievements (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    achievement_id uuid NOT NULL,
    earned_at timestamp with time zone DEFAULT now() NOT NULL,
    display_id bigint NOT NULL
);


--
-- Name: volunteer_achievements_display_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.volunteer_achievements_display_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: volunteer_achievements_display_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.volunteer_achievements_display_id_seq OWNED BY public.volunteer_achievements.display_id;


--
-- Name: volunteer_profiles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.volunteer_profiles (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    city text,
    phone text,
    bio text DEFAULT ''::text NOT NULL,
    total_hours numeric(8,2) DEFAULT 0 NOT NULL,
    points integer DEFAULT 0 NOT NULL,
    level integer DEFAULT 1 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    status public.volunteer_status DEFAULT 'active'::public.volunteer_status NOT NULL,
    interests text DEFAULT ''::text NOT NULL,
    display_id bigint NOT NULL
);


--
-- Name: volunteer_profiles_display_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.volunteer_profiles_display_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: volunteer_profiles_display_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.volunteer_profiles_display_id_seq OWNED BY public.volunteer_profiles.display_id;


--
-- Name: volunteer_skills; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.volunteer_skills (
    volunteer_profile_id uuid NOT NULL,
    skill_id uuid NOT NULL,
    display_id bigint NOT NULL
);


--
-- Name: volunteer_skills_display_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.volunteer_skills_display_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: volunteer_skills_display_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.volunteer_skills_display_id_seq OWNED BY public.volunteer_skills.display_id;


--
-- Name: achievement_rules display_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.achievement_rules ALTER COLUMN display_id SET DEFAULT nextval('public.achievement_rules_display_id_seq'::regclass);


--
-- Name: achievements display_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.achievements ALTER COLUMN display_id SET DEFAULT nextval('public.achievements_display_id_seq'::regclass);


--
-- Name: audit_log display_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.audit_log ALTER COLUMN display_id SET DEFAULT nextval('public.audit_log_display_id_seq'::regclass);


--
-- Name: certificates display_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.certificates ALTER COLUMN display_id SET DEFAULT nextval('public.certificates_display_id_seq'::regclass);


--
-- Name: email_verification_tokens display_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.email_verification_tokens ALTER COLUMN display_id SET DEFAULT nextval('public.email_verification_tokens_display_id_seq'::regclass);


--
-- Name: event_applications display_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.event_applications ALTER COLUMN display_id SET DEFAULT nextval('public.event_applications_display_id_seq'::regclass);


--
-- Name: event_attendance display_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.event_attendance ALTER COLUMN display_id SET DEFAULT nextval('public.event_attendance_display_id_seq'::regclass);


--
-- Name: event_feedback display_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.event_feedback ALTER COLUMN display_id SET DEFAULT nextval('public.event_feedback_display_id_seq'::regclass);


--
-- Name: event_shifts display_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.event_shifts ALTER COLUMN display_id SET DEFAULT nextval('public.event_shifts_display_id_seq'::regclass);


--
-- Name: event_templates display_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.event_templates ALTER COLUMN display_id SET DEFAULT nextval('public.event_templates_display_id_seq'::regclass);


--
-- Name: events display_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.events ALTER COLUMN display_id SET DEFAULT nextval('public.events_display_id_seq'::regclass);


--
-- Name: field_checkins display_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.field_checkins ALTER COLUMN display_id SET DEFAULT nextval('public.field_checkins_display_id_seq'::regclass);


--
-- Name: field_qr_tokens display_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.field_qr_tokens ALTER COLUMN display_id SET DEFAULT nextval('public.field_qr_tokens_display_id_seq'::regclass);


--
-- Name: knowledge_articles display_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.knowledge_articles ALTER COLUMN display_id SET DEFAULT nextval('public.knowledge_articles_display_id_seq'::regclass);


--
-- Name: knowledge_categories display_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.knowledge_categories ALTER COLUMN display_id SET DEFAULT nextval('public.knowledge_categories_display_id_seq'::regclass);


--
-- Name: news display_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.news ALTER COLUMN display_id SET DEFAULT nextval('public.news_display_id_seq'::regclass);


--
-- Name: news_categories display_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.news_categories ALTER COLUMN display_id SET DEFAULT nextval('public.news_categories_display_id_seq'::regclass);


--
-- Name: notifications display_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications ALTER COLUMN display_id SET DEFAULT nextval('public.notifications_display_id_seq'::regclass);


--
-- Name: organization_members display_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.organization_members ALTER COLUMN display_id SET DEFAULT nextval('public.organization_members_display_id_seq'::regclass);


--
-- Name: organizations display_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.organizations ALTER COLUMN display_id SET DEFAULT nextval('public.organizations_display_id_seq'::regclass);


--
-- Name: password_reset_tokens display_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.password_reset_tokens ALTER COLUMN display_id SET DEFAULT nextval('public.password_reset_tokens_display_id_seq'::regclass);


--
-- Name: points_transactions display_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.points_transactions ALTER COLUMN display_id SET DEFAULT nextval('public.points_transactions_display_id_seq'::regclass);


--
-- Name: refresh_sessions display_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.refresh_sessions ALTER COLUMN display_id SET DEFAULT nextval('public.refresh_sessions_display_id_seq'::regclass);


--
-- Name: revoked_access_tokens display_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.revoked_access_tokens ALTER COLUMN display_id SET DEFAULT nextval('public.revoked_access_tokens_display_id_seq'::regclass);


--
-- Name: skills display_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.skills ALTER COLUMN display_id SET DEFAULT nextval('public.skills_display_id_seq'::regclass);


--
-- Name: system_roles display_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_roles ALTER COLUMN display_id SET DEFAULT nextval('public.system_roles_display_id_seq'::regclass);


--
-- Name: task_assignments display_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_assignments ALTER COLUMN display_id SET DEFAULT nextval('public.task_assignments_display_id_seq'::regclass);


--
-- Name: task_attachments display_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_attachments ALTER COLUMN display_id SET DEFAULT nextval('public.task_attachments_display_id_seq'::regclass);


--
-- Name: task_comments display_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_comments ALTER COLUMN display_id SET DEFAULT nextval('public.task_comments_display_id_seq'::regclass);


--
-- Name: task_status_history display_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_status_history ALTER COLUMN display_id SET DEFAULT nextval('public.task_status_history_display_id_seq'::regclass);


--
-- Name: task_time_entries display_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_time_entries ALTER COLUMN display_id SET DEFAULT nextval('public.task_time_entries_display_id_seq'::regclass);


--
-- Name: tasks display_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tasks ALTER COLUMN display_id SET DEFAULT nextval('public.tasks_display_id_seq'::regclass);


--
-- Name: tenant_settings display_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_settings ALTER COLUMN display_id SET DEFAULT nextval('public.tenant_settings_display_id_seq'::regclass);


--
-- Name: time_entries display_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.time_entries ALTER COLUMN display_id SET DEFAULT nextval('public.time_entries_display_id_seq'::regclass);


--
-- Name: user_invitations display_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_invitations ALTER COLUMN display_id SET DEFAULT nextval('public.user_invitations_display_id_seq'::regclass);


--
-- Name: user_system_roles display_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_system_roles ALTER COLUMN display_id SET DEFAULT nextval('public.user_system_roles_display_id_seq'::regclass);


--
-- Name: users display_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users ALTER COLUMN display_id SET DEFAULT nextval('public.users_display_id_seq'::regclass);


--
-- Name: volunteer_achievements display_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.volunteer_achievements ALTER COLUMN display_id SET DEFAULT nextval('public.volunteer_achievements_display_id_seq'::regclass);


--
-- Name: volunteer_profiles display_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.volunteer_profiles ALTER COLUMN display_id SET DEFAULT nextval('public.volunteer_profiles_display_id_seq'::regclass);


--
-- Name: volunteer_skills display_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.volunteer_skills ALTER COLUMN display_id SET DEFAULT nextval('public.volunteer_skills_display_id_seq'::regclass);


--
-- Name: achievement_rules achievement_rules_achievement_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.achievement_rules
    ADD CONSTRAINT achievement_rules_achievement_id_key UNIQUE (achievement_id);


--
-- Name: achievement_rules achievement_rules_code_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.achievement_rules
    ADD CONSTRAINT achievement_rules_code_key UNIQUE (code);


--
-- Name: achievement_rules achievement_rules_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.achievement_rules
    ADD CONSTRAINT achievement_rules_pkey PRIMARY KEY (id);


--
-- Name: achievements achievements_code_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.achievements
    ADD CONSTRAINT achievements_code_key UNIQUE (code);


--
-- Name: achievements achievements_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.achievements
    ADD CONSTRAINT achievements_pkey PRIMARY KEY (id);


--
-- Name: audit_log audit_log_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.audit_log
    ADD CONSTRAINT audit_log_pkey PRIMARY KEY (id);


--
-- Name: certificates certificates_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.certificates
    ADD CONSTRAINT certificates_pkey PRIMARY KEY (id);


--
-- Name: certificates certificates_verify_code_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.certificates
    ADD CONSTRAINT certificates_verify_code_key UNIQUE (verify_code);


--
-- Name: email_verification_tokens email_verification_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.email_verification_tokens
    ADD CONSTRAINT email_verification_tokens_pkey PRIMARY KEY (id);


--
-- Name: email_verification_tokens email_verification_tokens_token_hash_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.email_verification_tokens
    ADD CONSTRAINT email_verification_tokens_token_hash_key UNIQUE (token_hash);


--
-- Name: event_applications event_applications_event_id_user_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.event_applications
    ADD CONSTRAINT event_applications_event_id_user_id_key UNIQUE (event_id, user_id);


--
-- Name: event_applications event_applications_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.event_applications
    ADD CONSTRAINT event_applications_pkey PRIMARY KEY (id);


--
-- Name: event_attendance event_attendance_event_id_user_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.event_attendance
    ADD CONSTRAINT event_attendance_event_id_user_id_key UNIQUE (event_id, user_id);


--
-- Name: event_attendance event_attendance_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.event_attendance
    ADD CONSTRAINT event_attendance_pkey PRIMARY KEY (id);


--
-- Name: event_feedback event_feedback_event_id_user_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.event_feedback
    ADD CONSTRAINT event_feedback_event_id_user_id_key UNIQUE (event_id, user_id);


--
-- Name: event_feedback event_feedback_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.event_feedback
    ADD CONSTRAINT event_feedback_pkey PRIMARY KEY (id);


--
-- Name: event_shifts event_shifts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.event_shifts
    ADD CONSTRAINT event_shifts_pkey PRIMARY KEY (id);


--
-- Name: event_templates event_templates_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.event_templates
    ADD CONSTRAINT event_templates_pkey PRIMARY KEY (id);


--
-- Name: events events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.events
    ADD CONSTRAINT events_pkey PRIMARY KEY (id);


--
-- Name: field_checkins field_checkins_event_id_user_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.field_checkins
    ADD CONSTRAINT field_checkins_event_id_user_id_key UNIQUE (event_id, user_id);


--
-- Name: field_checkins field_checkins_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.field_checkins
    ADD CONSTRAINT field_checkins_pkey PRIMARY KEY (id);


--
-- Name: field_qr_tokens field_qr_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.field_qr_tokens
    ADD CONSTRAINT field_qr_tokens_pkey PRIMARY KEY (id);


--
-- Name: field_qr_tokens field_qr_tokens_token_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.field_qr_tokens
    ADD CONSTRAINT field_qr_tokens_token_key UNIQUE (token);


--
-- Name: knowledge_articles knowledge_articles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.knowledge_articles
    ADD CONSTRAINT knowledge_articles_pkey PRIMARY KEY (id);


--
-- Name: knowledge_articles knowledge_articles_slug_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.knowledge_articles
    ADD CONSTRAINT knowledge_articles_slug_key UNIQUE (slug);


--
-- Name: knowledge_categories knowledge_categories_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.knowledge_categories
    ADD CONSTRAINT knowledge_categories_name_key UNIQUE (name);


--
-- Name: knowledge_categories knowledge_categories_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.knowledge_categories
    ADD CONSTRAINT knowledge_categories_pkey PRIMARY KEY (id);


--
-- Name: knowledge_categories knowledge_categories_slug_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.knowledge_categories
    ADD CONSTRAINT knowledge_categories_slug_key UNIQUE (slug);


--
-- Name: news_categories news_categories_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.news_categories
    ADD CONSTRAINT news_categories_name_key UNIQUE (name);


--
-- Name: news_categories news_categories_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.news_categories
    ADD CONSTRAINT news_categories_pkey PRIMARY KEY (id);


--
-- Name: news_categories news_categories_slug_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.news_categories
    ADD CONSTRAINT news_categories_slug_key UNIQUE (slug);


--
-- Name: news news_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.news
    ADD CONSTRAINT news_pkey PRIMARY KEY (id);


--
-- Name: news news_slug_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.news
    ADD CONSTRAINT news_slug_key UNIQUE (slug);


--
-- Name: notifications notifications_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_pkey PRIMARY KEY (id);


--
-- Name: organization_members organization_members_organization_id_user_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.organization_members
    ADD CONSTRAINT organization_members_organization_id_user_id_key UNIQUE (organization_id, user_id);


--
-- Name: organization_members organization_members_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.organization_members
    ADD CONSTRAINT organization_members_pkey PRIMARY KEY (id);


--
-- Name: organizations organizations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.organizations
    ADD CONSTRAINT organizations_pkey PRIMARY KEY (id);


--
-- Name: organizations organizations_slug_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.organizations
    ADD CONSTRAINT organizations_slug_key UNIQUE (slug);


--
-- Name: password_reset_tokens password_reset_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.password_reset_tokens
    ADD CONSTRAINT password_reset_tokens_pkey PRIMARY KEY (id);


--
-- Name: password_reset_tokens password_reset_tokens_token_hash_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.password_reset_tokens
    ADD CONSTRAINT password_reset_tokens_token_hash_key UNIQUE (token_hash);


--
-- Name: points_transactions points_transactions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.points_transactions
    ADD CONSTRAINT points_transactions_pkey PRIMARY KEY (id);


--
-- Name: points_transactions points_transactions_user_id_achievement_id_source_type_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.points_transactions
    ADD CONSTRAINT points_transactions_user_id_achievement_id_source_type_key UNIQUE (user_id, achievement_id, source_type);


--
-- Name: profile_field_definitions profile_field_definitions_organization_id_code_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.profile_field_definitions
    ADD CONSTRAINT profile_field_definitions_organization_id_code_key UNIQUE (organization_id, code);


--
-- Name: profile_field_definitions profile_field_definitions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.profile_field_definitions
    ADD CONSTRAINT profile_field_definitions_pkey PRIMARY KEY (id);


--
-- Name: profile_field_groups profile_field_groups_organization_id_code_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.profile_field_groups
    ADD CONSTRAINT profile_field_groups_organization_id_code_key UNIQUE (organization_id, code);


--
-- Name: profile_field_groups profile_field_groups_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.profile_field_groups
    ADD CONSTRAINT profile_field_groups_pkey PRIMARY KEY (id);


--
-- Name: profile_field_options profile_field_options_field_id_value_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.profile_field_options
    ADD CONSTRAINT profile_field_options_field_id_value_key UNIQUE (field_id, value);


--
-- Name: profile_field_options profile_field_options_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.profile_field_options
    ADD CONSTRAINT profile_field_options_pkey PRIMARY KEY (id);


--
-- Name: profile_field_values profile_field_values_organization_id_user_id_field_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.profile_field_values
    ADD CONSTRAINT profile_field_values_organization_id_user_id_field_id_key UNIQUE (organization_id, user_id, field_id);


--
-- Name: profile_field_values profile_field_values_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.profile_field_values
    ADD CONSTRAINT profile_field_values_pkey PRIMARY KEY (id);


--
-- Name: refresh_sessions refresh_sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.refresh_sessions
    ADD CONSTRAINT refresh_sessions_pkey PRIMARY KEY (id);


--
-- Name: refresh_sessions refresh_sessions_token_hash_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.refresh_sessions
    ADD CONSTRAINT refresh_sessions_token_hash_key UNIQUE (token_hash);


--
-- Name: revoked_access_tokens revoked_access_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.revoked_access_tokens
    ADD CONSTRAINT revoked_access_tokens_pkey PRIMARY KEY (jti);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (name);


--
-- Name: seed_migrations seed_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.seed_migrations
    ADD CONSTRAINT seed_migrations_pkey PRIMARY KEY (name);


--
-- Name: skills skills_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.skills
    ADD CONSTRAINT skills_name_key UNIQUE (name);


--
-- Name: skills skills_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.skills
    ADD CONSTRAINT skills_pkey PRIMARY KEY (id);


--
-- Name: system_roles system_roles_code_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_roles
    ADD CONSTRAINT system_roles_code_key UNIQUE (code);


--
-- Name: system_roles system_roles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_roles
    ADD CONSTRAINT system_roles_pkey PRIMARY KEY (id);


--
-- Name: task_assignments task_assignments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_assignments
    ADD CONSTRAINT task_assignments_pkey PRIMARY KEY (id);


--
-- Name: task_assignments task_assignments_task_id_user_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_assignments
    ADD CONSTRAINT task_assignments_task_id_user_id_key UNIQUE (task_id, user_id);


--
-- Name: task_attachments task_attachments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_attachments
    ADD CONSTRAINT task_attachments_pkey PRIMARY KEY (id);


--
-- Name: task_comments task_comments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_comments
    ADD CONSTRAINT task_comments_pkey PRIMARY KEY (id);


--
-- Name: task_status_history task_status_history_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_status_history
    ADD CONSTRAINT task_status_history_pkey PRIMARY KEY (id);


--
-- Name: task_time_entries task_time_entries_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_time_entries
    ADD CONSTRAINT task_time_entries_pkey PRIMARY KEY (id);


--
-- Name: tasks tasks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_pkey PRIMARY KEY (id);


--
-- Name: tenant_settings tenant_settings_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_settings
    ADD CONSTRAINT tenant_settings_pkey PRIMARY KEY (id);


--
-- Name: tenant_settings tenant_settings_singleton_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_settings
    ADD CONSTRAINT tenant_settings_singleton_unique UNIQUE (singleton_key);


--
-- Name: time_entries time_entries_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.time_entries
    ADD CONSTRAINT time_entries_pkey PRIMARY KEY (id);


--
-- Name: user_invitations user_invitations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_invitations
    ADD CONSTRAINT user_invitations_pkey PRIMARY KEY (id);


--
-- Name: user_invitations user_invitations_token_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_invitations
    ADD CONSTRAINT user_invitations_token_key UNIQUE (token);


--
-- Name: user_system_roles user_system_roles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_system_roles
    ADD CONSTRAINT user_system_roles_pkey PRIMARY KEY (user_id, role_id);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: volunteer_achievements volunteer_achievements_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.volunteer_achievements
    ADD CONSTRAINT volunteer_achievements_pkey PRIMARY KEY (id);


--
-- Name: volunteer_achievements volunteer_achievements_user_id_achievement_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.volunteer_achievements
    ADD CONSTRAINT volunteer_achievements_user_id_achievement_id_key UNIQUE (user_id, achievement_id);


--
-- Name: volunteer_profiles volunteer_profiles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.volunteer_profiles
    ADD CONSTRAINT volunteer_profiles_pkey PRIMARY KEY (id);


--
-- Name: volunteer_profiles volunteer_profiles_user_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.volunteer_profiles
    ADD CONSTRAINT volunteer_profiles_user_id_key UNIQUE (user_id);


--
-- Name: volunteer_skills volunteer_skills_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.volunteer_skills
    ADD CONSTRAINT volunteer_skills_pkey PRIMARY KEY (volunteer_profile_id, skill_id);


--
-- Name: achievement_rules_display_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX achievement_rules_display_id_key ON public.achievement_rules USING btree (display_id);


--
-- Name: achievements_display_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX achievements_display_id_key ON public.achievements USING btree (display_id);


--
-- Name: audit_log_display_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX audit_log_display_id_key ON public.audit_log USING btree (display_id);


--
-- Name: certificates_display_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX certificates_display_id_key ON public.certificates USING btree (display_id);


--
-- Name: email_verification_tokens_display_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX email_verification_tokens_display_id_key ON public.email_verification_tokens USING btree (display_id);


--
-- Name: event_applications_display_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX event_applications_display_id_key ON public.event_applications USING btree (display_id);


--
-- Name: event_attendance_display_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX event_attendance_display_id_key ON public.event_attendance USING btree (display_id);


--
-- Name: event_feedback_display_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX event_feedback_display_id_key ON public.event_feedback USING btree (display_id);


--
-- Name: event_shifts_display_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX event_shifts_display_id_key ON public.event_shifts USING btree (display_id);


--
-- Name: event_templates_display_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX event_templates_display_id_key ON public.event_templates USING btree (display_id);


--
-- Name: events_display_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX events_display_id_key ON public.events USING btree (display_id);


--
-- Name: field_checkins_display_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX field_checkins_display_id_key ON public.field_checkins USING btree (display_id);


--
-- Name: field_qr_tokens_display_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX field_qr_tokens_display_id_key ON public.field_qr_tokens USING btree (display_id);


--
-- Name: idx_audit_log_action; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_log_action ON public.audit_log USING btree (action);


--
-- Name: idx_audit_log_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_log_created_at ON public.audit_log USING btree (created_at DESC);


--
-- Name: idx_audit_log_entity_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_log_entity_type ON public.audit_log USING btree (entity_type);


--
-- Name: idx_audit_log_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_log_user_id ON public.audit_log USING btree (user_id);


--
-- Name: idx_certificates_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_certificates_user_id ON public.certificates USING btree (user_id);


--
-- Name: idx_certificates_verify_code; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_certificates_verify_code ON public.certificates USING btree (verify_code);


--
-- Name: idx_email_verification_tokens_token_hash; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_email_verification_tokens_token_hash ON public.email_verification_tokens USING btree (token_hash);


--
-- Name: idx_event_applications_event_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_event_applications_event_id ON public.event_applications USING btree (event_id);


--
-- Name: idx_event_attendance_event_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_event_attendance_event_id ON public.event_attendance USING btree (event_id);


--
-- Name: idx_event_templates_organization_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_event_templates_organization_id ON public.event_templates USING btree (organization_id);


--
-- Name: idx_events_organization_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_events_organization_id ON public.events USING btree (organization_id);


--
-- Name: idx_field_checkins_event_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_field_checkins_event_id ON public.field_checkins USING btree (event_id);


--
-- Name: idx_field_checkins_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_field_checkins_user_id ON public.field_checkins USING btree (user_id);


--
-- Name: idx_field_qr_tokens_event_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_field_qr_tokens_event_id ON public.field_qr_tokens USING btree (event_id);


--
-- Name: idx_field_qr_tokens_token; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_field_qr_tokens_token ON public.field_qr_tokens USING btree (token);


--
-- Name: idx_knowledge_articles_category_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_knowledge_articles_category_id ON public.knowledge_articles USING btree (category_id);


--
-- Name: idx_knowledge_articles_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_knowledge_articles_status ON public.knowledge_articles USING btree (status);


--
-- Name: idx_knowledge_articles_title; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_knowledge_articles_title ON public.knowledge_articles USING gin (to_tsvector('simple'::regconfig, ((((title || ' '::text) || summary) || ' '::text) || content_html)));


--
-- Name: idx_news_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_news_created_at ON public.news USING btree (created_at DESC);


--
-- Name: idx_news_organization_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_news_organization_id ON public.news USING btree (organization_id);


--
-- Name: idx_news_scheduled_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_news_scheduled_at ON public.news USING btree (scheduled_at);


--
-- Name: idx_news_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_news_status ON public.news USING btree (status);


--
-- Name: idx_notifications_dedupe; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_notifications_dedupe ON public.notifications USING btree (user_id, type, entity_type, entity_id);


--
-- Name: idx_notifications_user_created; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_notifications_user_created ON public.notifications USING btree (user_id, is_read, created_at DESC);


--
-- Name: idx_organization_members_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_organization_members_user_id ON public.organization_members USING btree (user_id);


--
-- Name: idx_organizations_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_organizations_deleted_at ON public.organizations USING btree (deleted_at);


--
-- Name: idx_organizations_name_search; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_organizations_name_search ON public.organizations USING gin (to_tsvector('simple'::regconfig, ((((name || ' '::text) || slug) || ' '::text) || description)));


--
-- Name: idx_password_reset_tokens_token_hash; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_password_reset_tokens_token_hash ON public.password_reset_tokens USING btree (token_hash);


--
-- Name: idx_points_transactions_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_points_transactions_user_id ON public.points_transactions USING btree (user_id);


--
-- Name: idx_profile_fields_group; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_profile_fields_group ON public.profile_field_definitions USING btree (group_id, sort_order);


--
-- Name: idx_profile_groups_organization; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_profile_groups_organization ON public.profile_field_groups USING btree (organization_id, sort_order);


--
-- Name: idx_profile_values_user; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_profile_values_user ON public.profile_field_values USING btree (organization_id, user_id);


--
-- Name: idx_refresh_sessions_token_hash; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_refresh_sessions_token_hash ON public.refresh_sessions USING btree (token_hash);


--
-- Name: idx_refresh_sessions_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_refresh_sessions_user_id ON public.refresh_sessions USING btree (user_id);


--
-- Name: idx_revoked_access_tokens_expires_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_revoked_access_tokens_expires_at ON public.revoked_access_tokens USING btree (expires_at);


--
-- Name: idx_task_comments_task_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_task_comments_task_id ON public.task_comments USING btree (task_id);


--
-- Name: idx_task_status_history_task_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_task_status_history_task_id ON public.task_status_history USING btree (task_id);


--
-- Name: idx_tasks_organization_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tasks_organization_id ON public.tasks USING btree (organization_id);


--
-- Name: idx_time_entries_organization_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_time_entries_organization_id ON public.time_entries USING btree (organization_id);


--
-- Name: idx_time_entries_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_time_entries_status ON public.time_entries USING btree (status);


--
-- Name: idx_time_entries_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_time_entries_user_id ON public.time_entries USING btree (user_id);


--
-- Name: idx_user_invitations_email; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_user_invitations_email ON public.user_invitations USING btree (lower(email));


--
-- Name: idx_user_invitations_token; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_user_invitations_token ON public.user_invitations USING btree (token);


--
-- Name: idx_volunteer_achievements_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_volunteer_achievements_user_id ON public.volunteer_achievements USING btree (user_id);


--
-- Name: idx_volunteer_profiles_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_volunteer_profiles_status ON public.volunteer_profiles USING btree (status);


--
-- Name: knowledge_articles_display_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX knowledge_articles_display_id_key ON public.knowledge_articles USING btree (display_id);


--
-- Name: knowledge_categories_display_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX knowledge_categories_display_id_key ON public.knowledge_categories USING btree (display_id);


--
-- Name: news_categories_display_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX news_categories_display_id_key ON public.news_categories USING btree (display_id);


--
-- Name: news_display_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX news_display_id_key ON public.news USING btree (display_id);


--
-- Name: notifications_display_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX notifications_display_id_key ON public.notifications USING btree (display_id);


--
-- Name: organization_members_display_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX organization_members_display_id_key ON public.organization_members USING btree (display_id);


--
-- Name: organizations_display_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX organizations_display_id_key ON public.organizations USING btree (display_id);


--
-- Name: password_reset_tokens_display_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX password_reset_tokens_display_id_key ON public.password_reset_tokens USING btree (display_id);


--
-- Name: points_transactions_display_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX points_transactions_display_id_key ON public.points_transactions USING btree (display_id);


--
-- Name: refresh_sessions_display_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX refresh_sessions_display_id_key ON public.refresh_sessions USING btree (display_id);


--
-- Name: revoked_access_tokens_display_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX revoked_access_tokens_display_id_key ON public.revoked_access_tokens USING btree (display_id);


--
-- Name: skills_display_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX skills_display_id_key ON public.skills USING btree (display_id);


--
-- Name: system_roles_display_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX system_roles_display_id_key ON public.system_roles USING btree (display_id);


--
-- Name: task_assignments_display_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX task_assignments_display_id_key ON public.task_assignments USING btree (display_id);


--
-- Name: task_attachments_display_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX task_attachments_display_id_key ON public.task_attachments USING btree (display_id);


--
-- Name: task_comments_display_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX task_comments_display_id_key ON public.task_comments USING btree (display_id);


--
-- Name: task_status_history_display_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX task_status_history_display_id_key ON public.task_status_history USING btree (display_id);


--
-- Name: task_time_entries_display_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX task_time_entries_display_id_key ON public.task_time_entries USING btree (display_id);


--
-- Name: tasks_display_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX tasks_display_id_key ON public.tasks USING btree (display_id);


--
-- Name: tenant_settings_display_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX tenant_settings_display_id_key ON public.tenant_settings USING btree (display_id);


--
-- Name: time_entries_display_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX time_entries_display_id_key ON public.time_entries USING btree (display_id);


--
-- Name: user_invitations_display_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX user_invitations_display_id_key ON public.user_invitations USING btree (display_id);


--
-- Name: user_system_roles_display_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX user_system_roles_display_id_key ON public.user_system_roles USING btree (display_id);


--
-- Name: users_display_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX users_display_id_key ON public.users USING btree (display_id);


--
-- Name: volunteer_achievements_display_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX volunteer_achievements_display_id_key ON public.volunteer_achievements USING btree (display_id);


--
-- Name: volunteer_profiles_display_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX volunteer_profiles_display_id_key ON public.volunteer_profiles USING btree (display_id);


--
-- Name: volunteer_skills_display_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX volunteer_skills_display_id_key ON public.volunteer_skills USING btree (display_id);


--
-- Name: volunteer_achievements trg_achievement_notification; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_achievement_notification AFTER INSERT ON public.volunteer_achievements FOR EACH ROW EXECUTE FUNCTION public.create_achievement_notification();


--
-- Name: achievement_rules achievement_rules_achievement_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.achievement_rules
    ADD CONSTRAINT achievement_rules_achievement_id_fkey FOREIGN KEY (achievement_id) REFERENCES public.achievements(id) ON DELETE CASCADE;


--
-- Name: audit_log audit_log_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.audit_log
    ADD CONSTRAINT audit_log_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: certificates certificates_issued_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.certificates
    ADD CONSTRAINT certificates_issued_by_fkey FOREIGN KEY (issued_by) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: certificates certificates_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.certificates
    ADD CONSTRAINT certificates_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE SET NULL;


--
-- Name: certificates certificates_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.certificates
    ADD CONSTRAINT certificates_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: email_verification_tokens email_verification_tokens_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.email_verification_tokens
    ADD CONSTRAINT email_verification_tokens_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: event_applications event_applications_event_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.event_applications
    ADD CONSTRAINT event_applications_event_id_fkey FOREIGN KEY (event_id) REFERENCES public.events(id) ON DELETE CASCADE;


--
-- Name: event_applications event_applications_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.event_applications
    ADD CONSTRAINT event_applications_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: event_attendance event_attendance_event_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.event_attendance
    ADD CONSTRAINT event_attendance_event_id_fkey FOREIGN KEY (event_id) REFERENCES public.events(id) ON DELETE CASCADE;


--
-- Name: event_attendance event_attendance_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.event_attendance
    ADD CONSTRAINT event_attendance_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: event_feedback event_feedback_event_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.event_feedback
    ADD CONSTRAINT event_feedback_event_id_fkey FOREIGN KEY (event_id) REFERENCES public.events(id) ON DELETE CASCADE;


--
-- Name: event_feedback event_feedback_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.event_feedback
    ADD CONSTRAINT event_feedback_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: event_shifts event_shifts_event_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.event_shifts
    ADD CONSTRAINT event_shifts_event_id_fkey FOREIGN KEY (event_id) REFERENCES public.events(id) ON DELETE CASCADE;


--
-- Name: event_templates event_templates_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.event_templates
    ADD CONSTRAINT event_templates_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: event_templates event_templates_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.event_templates
    ADD CONSTRAINT event_templates_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: events events_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.events
    ADD CONSTRAINT events_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: events events_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.events
    ADD CONSTRAINT events_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: field_checkins field_checkins_event_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.field_checkins
    ADD CONSTRAINT field_checkins_event_id_fkey FOREIGN KEY (event_id) REFERENCES public.events(id) ON DELETE CASCADE;


--
-- Name: field_checkins field_checkins_qr_token_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.field_checkins
    ADD CONSTRAINT field_checkins_qr_token_id_fkey FOREIGN KEY (qr_token_id) REFERENCES public.field_qr_tokens(id) ON DELETE SET NULL;


--
-- Name: field_checkins field_checkins_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.field_checkins
    ADD CONSTRAINT field_checkins_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: field_qr_tokens field_qr_tokens_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.field_qr_tokens
    ADD CONSTRAINT field_qr_tokens_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: field_qr_tokens field_qr_tokens_event_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.field_qr_tokens
    ADD CONSTRAINT field_qr_tokens_event_id_fkey FOREIGN KEY (event_id) REFERENCES public.events(id) ON DELETE CASCADE;


--
-- Name: knowledge_articles knowledge_articles_author_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.knowledge_articles
    ADD CONSTRAINT knowledge_articles_author_id_fkey FOREIGN KEY (author_id) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: knowledge_articles knowledge_articles_category_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.knowledge_articles
    ADD CONSTRAINT knowledge_articles_category_id_fkey FOREIGN KEY (category_id) REFERENCES public.knowledge_categories(id) ON DELETE SET NULL;


--
-- Name: news news_author_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.news
    ADD CONSTRAINT news_author_id_fkey FOREIGN KEY (author_id) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: news news_category_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.news
    ADD CONSTRAINT news_category_id_fkey FOREIGN KEY (category_id) REFERENCES public.news_categories(id) ON DELETE SET NULL;


--
-- Name: news news_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.news
    ADD CONSTRAINT news_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE SET NULL;


--
-- Name: notifications notifications_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: organization_members organization_members_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.organization_members
    ADD CONSTRAINT organization_members_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: organization_members organization_members_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.organization_members
    ADD CONSTRAINT organization_members_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: password_reset_tokens password_reset_tokens_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.password_reset_tokens
    ADD CONSTRAINT password_reset_tokens_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: points_transactions points_transactions_achievement_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.points_transactions
    ADD CONSTRAINT points_transactions_achievement_id_fkey FOREIGN KEY (achievement_id) REFERENCES public.achievements(id) ON DELETE SET NULL;


--
-- Name: points_transactions points_transactions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.points_transactions
    ADD CONSTRAINT points_transactions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: profile_field_definitions profile_field_definitions_group_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.profile_field_definitions
    ADD CONSTRAINT profile_field_definitions_group_id_fkey FOREIGN KEY (group_id) REFERENCES public.profile_field_groups(id) ON DELETE CASCADE;


--
-- Name: profile_field_definitions profile_field_definitions_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.profile_field_definitions
    ADD CONSTRAINT profile_field_definitions_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: profile_field_groups profile_field_groups_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.profile_field_groups
    ADD CONSTRAINT profile_field_groups_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: profile_field_options profile_field_options_field_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.profile_field_options
    ADD CONSTRAINT profile_field_options_field_id_fkey FOREIGN KEY (field_id) REFERENCES public.profile_field_definitions(id) ON DELETE CASCADE;


--
-- Name: profile_field_values profile_field_values_field_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.profile_field_values
    ADD CONSTRAINT profile_field_values_field_id_fkey FOREIGN KEY (field_id) REFERENCES public.profile_field_definitions(id) ON DELETE CASCADE;


--
-- Name: profile_field_values profile_field_values_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.profile_field_values
    ADD CONSTRAINT profile_field_values_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: profile_field_values profile_field_values_updated_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.profile_field_values
    ADD CONSTRAINT profile_field_values_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: profile_field_values profile_field_values_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.profile_field_values
    ADD CONSTRAINT profile_field_values_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: refresh_sessions refresh_sessions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.refresh_sessions
    ADD CONSTRAINT refresh_sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: revoked_access_tokens revoked_access_tokens_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.revoked_access_tokens
    ADD CONSTRAINT revoked_access_tokens_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: task_assignments task_assignments_task_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_assignments
    ADD CONSTRAINT task_assignments_task_id_fkey FOREIGN KEY (task_id) REFERENCES public.tasks(id) ON DELETE CASCADE;


--
-- Name: task_assignments task_assignments_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_assignments
    ADD CONSTRAINT task_assignments_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: task_attachments task_attachments_task_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_attachments
    ADD CONSTRAINT task_attachments_task_id_fkey FOREIGN KEY (task_id) REFERENCES public.tasks(id) ON DELETE CASCADE;


--
-- Name: task_attachments task_attachments_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_attachments
    ADD CONSTRAINT task_attachments_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: task_comments task_comments_task_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_comments
    ADD CONSTRAINT task_comments_task_id_fkey FOREIGN KEY (task_id) REFERENCES public.tasks(id) ON DELETE CASCADE;


--
-- Name: task_comments task_comments_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_comments
    ADD CONSTRAINT task_comments_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: task_status_history task_status_history_changed_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_status_history
    ADD CONSTRAINT task_status_history_changed_by_fkey FOREIGN KEY (changed_by) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: task_status_history task_status_history_task_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_status_history
    ADD CONSTRAINT task_status_history_task_id_fkey FOREIGN KEY (task_id) REFERENCES public.tasks(id) ON DELETE CASCADE;


--
-- Name: task_time_entries task_time_entries_task_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_time_entries
    ADD CONSTRAINT task_time_entries_task_id_fkey FOREIGN KEY (task_id) REFERENCES public.tasks(id) ON DELETE CASCADE;


--
-- Name: task_time_entries task_time_entries_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_time_entries
    ADD CONSTRAINT task_time_entries_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: tasks tasks_completion_confirmed_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_completion_confirmed_by_fkey FOREIGN KEY (completion_confirmed_by) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: tasks tasks_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: tasks tasks_event_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_event_id_fkey FOREIGN KEY (event_id) REFERENCES public.events(id) ON DELETE SET NULL;


--
-- Name: tasks tasks_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: tenant_settings tenant_settings_default_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_settings
    ADD CONSTRAINT tenant_settings_default_organization_id_fkey FOREIGN KEY (default_organization_id) REFERENCES public.organizations(id) ON DELETE SET NULL;


--
-- Name: time_entries time_entries_event_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.time_entries
    ADD CONSTRAINT time_entries_event_id_fkey FOREIGN KEY (event_id) REFERENCES public.events(id) ON DELETE SET NULL;


--
-- Name: time_entries time_entries_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.time_entries
    ADD CONSTRAINT time_entries_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: time_entries time_entries_reviewed_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.time_entries
    ADD CONSTRAINT time_entries_reviewed_by_fkey FOREIGN KEY (reviewed_by) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: time_entries time_entries_task_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.time_entries
    ADD CONSTRAINT time_entries_task_id_fkey FOREIGN KEY (task_id) REFERENCES public.tasks(id) ON DELETE SET NULL;


--
-- Name: time_entries time_entries_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.time_entries
    ADD CONSTRAINT time_entries_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: user_invitations user_invitations_accepted_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_invitations
    ADD CONSTRAINT user_invitations_accepted_by_fkey FOREIGN KEY (accepted_by) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: user_invitations user_invitations_invited_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_invitations
    ADD CONSTRAINT user_invitations_invited_by_fkey FOREIGN KEY (invited_by) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: user_system_roles user_system_roles_role_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_system_roles
    ADD CONSTRAINT user_system_roles_role_id_fkey FOREIGN KEY (role_id) REFERENCES public.system_roles(id) ON DELETE CASCADE;


--
-- Name: user_system_roles user_system_roles_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_system_roles
    ADD CONSTRAINT user_system_roles_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: volunteer_achievements volunteer_achievements_achievement_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.volunteer_achievements
    ADD CONSTRAINT volunteer_achievements_achievement_id_fkey FOREIGN KEY (achievement_id) REFERENCES public.achievements(id) ON DELETE CASCADE;


--
-- Name: volunteer_achievements volunteer_achievements_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.volunteer_achievements
    ADD CONSTRAINT volunteer_achievements_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: volunteer_profiles volunteer_profiles_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.volunteer_profiles
    ADD CONSTRAINT volunteer_profiles_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: volunteer_skills volunteer_skills_skill_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.volunteer_skills
    ADD CONSTRAINT volunteer_skills_skill_id_fkey FOREIGN KEY (skill_id) REFERENCES public.skills(id) ON DELETE CASCADE;


--
-- Name: volunteer_skills volunteer_skills_volunteer_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.volunteer_skills
    ADD CONSTRAINT volunteer_skills_volunteer_profile_id_fkey FOREIGN KEY (volunteer_profile_id) REFERENCES public.volunteer_profiles(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

\unrestrict 5Uaf3Y0gCcDOfpwBX5Hjf4hrdLAk40jx15k9rfd9mrqtWSIzviQfDI9n9pT8v1b

