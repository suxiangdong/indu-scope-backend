ALTER TABLE payment_orders
    ADD COLUMN entitlement_start_at DATETIME(3) NULL AFTER paid_at,
    ADD COLUMN entitlement_end_at DATETIME(3) NULL AFTER entitlement_start_at,
    ADD COLUMN entitlement_revoked_at DATETIME(3) NULL AFTER entitlement_end_at,
    ADD KEY idx_payment_orders_subscription (subscription_id, status, id);

-- Reconstruct historical grants only where all order durations exactly explain
-- the subscription's recorded lifetime. Manually changed/unknown periods stay
-- NULL and are flagged for manual entitlement review after a refund.
UPDATE payment_orders target
JOIN (
    SELECT o.id, s.end_at AS subscription_end_at,
           SUM(CAST(o.duration_days_snapshot AS SIGNED) * 86400) OVER (
               PARTITION BY o.subscription_id ORDER BY COALESCE(o.paid_at, o.created_at) DESC, o.id DESC
               ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
           ) AS tail_seconds,
           CAST(o.duration_days_snapshot AS SIGNED) * 86400 AS grant_seconds
    FROM payment_orders o
    JOIN vip_subscriptions s ON s.id = o.subscription_id
    JOIN (
        SELECT subscription_id, SUM(CAST(duration_days_snapshot AS SIGNED) * 86400) AS total_seconds
        FROM payment_orders
        WHERE status = 'paid' AND subscription_id IS NOT NULL
        GROUP BY subscription_id
        HAVING MIN(COALESCE(duration_days_snapshot, 0)) > 0
    ) totals ON totals.subscription_id = s.id
    WHERE o.status = 'paid' AND s.start_at IS NOT NULL AND s.end_at IS NOT NULL
      AND s.end_at = TIMESTAMPADD(SECOND, totals.total_seconds, s.start_at)
) grants ON grants.id = target.id
SET target.entitlement_start_at = TIMESTAMPADD(SECOND, -CAST(grants.tail_seconds AS SIGNED), grants.subscription_end_at),
    target.entitlement_end_at = TIMESTAMPADD(SECOND, -CAST(grants.tail_seconds - grants.grant_seconds AS SIGNED), grants.subscription_end_at);

CREATE TABLE payment_refunds (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    refund_no VARCHAR(64) NOT NULL,
    order_id BIGINT UNSIGNED NOT NULL,
    order_no VARCHAR(64) NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    requested_by_admin_id BIGINT UNSIGNED NULL,
    reason VARCHAR(80) NOT NULL,
    amount_fen INT UNSIGNED NOT NULL,
    total_fen INT UNSIGNED NOT NULL,
    currency CHAR(3) NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'created',
    provider VARCHAR(32) NOT NULL DEFAULT 'wechatpay',
    provider_refund_id VARCHAR(64) NULL,
    payer_refund_fen INT UNSIGNED NULL,
    received_account VARCHAR(128) NULL,
    entitlement_status VARCHAR(16) NOT NULL DEFAULT 'pending',
    reclaimed_seconds BIGINT UNSIGNED NOT NULL DEFAULT 0,
    last_error VARCHAR(255) NULL,
    success_at DATETIME(3) NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    UNIQUE KEY uk_payment_refunds_no (refund_no),
    UNIQUE KEY uk_payment_refunds_provider_id (provider, provider_refund_id),
    KEY idx_payment_refunds_order (order_id, id),
    KEY idx_payment_refunds_status (status, created_at),
    CONSTRAINT fk_payment_refunds_order FOREIGN KEY (order_id) REFERENCES payment_orders (id) ON DELETE RESTRICT,
    CONSTRAINT fk_payment_refunds_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE RESTRICT,
    CONSTRAINT fk_payment_refunds_admin FOREIGN KEY (requested_by_admin_id) REFERENCES admin_users (id) ON DELETE SET NULL,
    CONSTRAINT chk_payment_refunds_full CHECK (amount_fen > 0 AND amount_fen = total_fen),
    CONSTRAINT chk_payment_refunds_status CHECK (status IN ('created', 'processing', 'success', 'closed', 'abnormal')),
    CONSTRAINT chk_payment_refunds_entitlement CHECK (entitlement_status IN ('pending', 'reclaimed', 'none', 'manual'))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
