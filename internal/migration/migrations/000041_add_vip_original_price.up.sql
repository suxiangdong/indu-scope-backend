ALTER TABLE vip_plans
    ADD COLUMN original_price_fen INT UNSIGNED NOT NULL DEFAULT 0 AFTER price_fen,
    ADD CONSTRAINT chk_vip_plans_original_price CHECK (original_price_fen = 0 OR original_price_fen >= price_fen);

ALTER TABLE payment_orders
    ADD COLUMN original_price_fen_snapshot INT UNSIGNED NULL AFTER plan_level_snapshot;

UPDATE payment_orders o
JOIN vip_plans p ON p.id = o.plan_id
SET o.original_price_fen_snapshot = p.original_price_fen;
