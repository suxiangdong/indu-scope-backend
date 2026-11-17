ALTER TABLE payment_orders
    DROP INDEX uk_payment_orders_provider_transaction,
    DROP COLUMN payer_openid,
    DROP COLUMN payment_mch_id,
    DROP COLUMN payment_app_id;
