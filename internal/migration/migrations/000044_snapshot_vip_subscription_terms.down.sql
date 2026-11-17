ALTER TABLE vip_subscriptions
    DROP CHECK chk_vip_subscription_plan_type,
    DROP COLUMN currency_snapshot,
    DROP COLUMN plan_type_snapshot;
