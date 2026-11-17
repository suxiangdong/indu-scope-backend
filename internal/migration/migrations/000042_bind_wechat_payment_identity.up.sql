ALTER TABLE payment_orders
    ADD COLUMN payment_app_id VARCHAR(32) NULL AFTER prepay_id,
    ADD COLUMN payment_mch_id VARCHAR(32) NULL AFTER payment_app_id,
    ADD COLUMN payer_openid VARCHAR(128) NULL AFTER payment_mch_id,
    ADD UNIQUE KEY uk_payment_orders_provider_transaction (provider, provider_transaction_id);
