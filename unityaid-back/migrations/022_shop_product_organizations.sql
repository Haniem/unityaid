UPDATE shop_products
SET organization_id = (SELECT id FROM organizations ORDER BY created_at LIMIT 1)
WHERE organization_id IS NULL
	AND EXISTS (SELECT 1 FROM organizations);

ALTER TABLE shop_products
	ALTER COLUMN organization_id SET NOT NULL;
