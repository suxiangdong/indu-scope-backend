DROP TABLE payment_refunds;
ALTER TABLE payment_orders
    DROP KEY idx_payment_orders_subscription,
    DROP COLUMN entitlement_revoked_at,
    DROP COLUMN entitlement_end_at,
    DROP COLUMN entitlement_start_at;
