CREATE OR REPLACE FUNCTION parse_period_from(value TEXT)
RETURNS TIMESTAMPTZ
LANGUAGE sql
STABLE
AS $$
	SELECT COALESCE(NULLIF(value, '')::timestamptz, now() - interval '30 days')
$$;

CREATE OR REPLACE FUNCTION parse_period_to(value TEXT)
RETURNS TIMESTAMPTZ
LANGUAGE sql
STABLE
AS $$
	SELECT COALESCE(NULLIF(value, '')::timestamptz, now())
$$;
