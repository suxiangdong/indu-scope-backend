ALTER TABLE payment_orders
    ADD COLUMN plan_code_snapshot VARCHAR(64) NULL AFTER plan_id,
    ADD COLUMN plan_name_snapshot VARCHAR(100) NULL AFTER plan_code_snapshot,
    ADD COLUMN plan_description_snapshot VARCHAR(500) NULL AFTER plan_name_snapshot,
    ADD COLUMN plan_level_snapshot INT UNSIGNED NULL AFTER plan_description_snapshot,
    ADD COLUMN plan_type_snapshot VARCHAR(16) NULL AFTER plan_level_snapshot,
    ADD COLUMN billing_cycle_snapshot VARCHAR(16) NULL AFTER plan_type_snapshot,
    ADD COLUMN duration_days_snapshot INT UNSIGNED NULL AFTER billing_cycle_snapshot,
    ADD COLUMN auto_renew_enabled_snapshot TINYINT UNSIGNED NULL AFTER duration_days_snapshot;

UPDATE payment_orders o
JOIN vip_plans p ON p.id = o.plan_id
SET o.plan_code_snapshot = p.code,
    o.plan_name_snapshot = p.name,
    o.plan_description_snapshot = p.description,
    o.plan_level_snapshot = p.level,
    o.plan_type_snapshot = p.plan_type,
    o.billing_cycle_snapshot = p.billing_cycle,
    o.duration_days_snapshot = p.duration_days,
    o.auto_renew_enabled_snapshot = p.auto_renew_enabled;
