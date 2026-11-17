ALTER TABLE payment_orders
    DROP COLUMN original_price_fen_snapshot;

ALTER TABLE vip_plans
    DROP CHECK chk_vip_plans_original_price,
    DROP COLUMN original_price_fen;
