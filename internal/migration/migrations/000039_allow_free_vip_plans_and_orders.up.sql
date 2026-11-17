ALTER TABLE vip_plans
    DROP CHECK chk_vip_plans_price,
    ADD CONSTRAINT chk_vip_plans_price CHECK (price_fen >= 0);

ALTER TABLE payment_orders
    DROP CHECK chk_payment_orders_amount,
    ADD CONSTRAINT chk_payment_orders_amount CHECK (amount_fen >= 0);
