ALTER TABLE volunteer_profiles
	ADD COLUMN IF NOT EXISTS coin_balance INTEGER NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS coin_transactions (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	amount INTEGER NOT NULL,
	type TEXT NOT NULL CHECK (type IN ('achievement', 'transfer_in', 'transfer_out', 'purchase', 'refund', 'demo_bonus')),
	description TEXT NOT NULL DEFAULT '',
	related_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
	source_type TEXT NOT NULL DEFAULT '',
	source_id UUID,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	UNIQUE (user_id, type, source_type, source_id)
);

CREATE TABLE IF NOT EXISTS shop_products (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	organization_id UUID REFERENCES organizations(id) ON DELETE SET NULL,
	name TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	price INTEGER NOT NULL CHECK (price > 0),
	stock INTEGER NOT NULL DEFAULT 0 CHECK (stock >= 0),
	image_url TEXT,
	is_active BOOLEAN NOT NULL DEFAULT true,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS shop_orders (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	organization_id UUID REFERENCES organizations(id) ON DELETE SET NULL,
	status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'cancelled')),
	total INTEGER NOT NULL CHECK (total >= 0),
	comment TEXT NOT NULL DEFAULT '',
	processed_by UUID REFERENCES users(id) ON DELETE SET NULL,
	processed_at TIMESTAMPTZ,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS shop_order_items (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	order_id UUID NOT NULL REFERENCES shop_orders(id) ON DELETE CASCADE,
	product_id UUID REFERENCES shop_products(id) ON DELETE SET NULL,
	product_name TEXT NOT NULL,
	price INTEGER NOT NULL CHECK (price > 0),
	quantity INTEGER NOT NULL CHECK (quantity > 0)
);

CREATE INDEX IF NOT EXISTS idx_coin_transactions_user ON coin_transactions(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_shop_orders_user ON shop_orders(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_shop_orders_status ON shop_orders(status, created_at DESC);

CREATE OR REPLACE FUNCTION award_coins_for_points()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
	IF NEW.points <= 0 THEN
		RETURN NEW;
	END IF;

	INSERT INTO coin_transactions (user_id, amount, type, description, source_type, source_id)
	VALUES (NEW.user_id, NEW.points, 'achievement', 'Монеты за достижение: ' || COALESCE(NEW.reason, ''), 'points_transaction', NEW.id)
	ON CONFLICT (user_id, type, source_type, source_id) DO NOTHING;

	UPDATE volunteer_profiles
	SET coin_balance = coin_balance + NEW.points,
		updated_at = now()
	WHERE user_id = NEW.user_id;

	RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_award_coins_for_points ON points_transactions;
CREATE TRIGGER trg_award_coins_for_points
AFTER INSERT ON points_transactions
FOR EACH ROW
EXECUTE FUNCTION award_coins_for_points();
