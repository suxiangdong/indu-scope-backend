-- Freeze the membership terms used for purchase eligibility. Catalog edits
-- must not change the type/currency of entitlements already purchased.
ALTER TABLE vip_subscriptions
    ADD COLUMN plan_type_snapshot VARCHAR(16) NOT NULL DEFAULT 'normal' AFTER plan_id,
    ADD COLUMN currency_snapshot CHAR(3) NOT NULL DEFAULT 'CNY' AFTER plan_type_snapshot;

UPDATE vip_subscriptions s
LEFT JOIN payment_orders o ON o.id = (
    SELECT MAX(candidate.id)
    FROM payment_orders candidate
    WHERE candidate.subscription_id = s.id AND candidate.status = 'paid'
)
LEFT JOIN vip_plans p ON p.id = s.plan_id
SET s.plan_type_snapshot = CASE
        WHEN o.plan_type_snapshot IN ('normal', 'auto_renew') THEN o.plan_type_snapshot
        WHEN s.auto_renew = 1 OR NULLIF(s.provider_contract_id, '') IS NOT NULL THEN 'auto_renew'
        ELSE COALESCE(p.plan_type, 'normal')
    END,
    s.currency_snapshot = COALESCE(NULLIF(o.currency, ''), NULLIF(p.currency, ''), 'CNY');

ALTER TABLE vip_subscriptions
    ADD CONSTRAINT chk_vip_subscription_plan_type CHECK (plan_type_snapshot IN ('normal', 'auto_renew'));

-- Historical vip_plans.level and payment_orders.plan_level_snapshot remain for
-- compatibility/audit only. Application APIs and business logic no longer use them.
