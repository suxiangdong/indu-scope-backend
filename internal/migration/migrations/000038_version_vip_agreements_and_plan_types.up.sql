ALTER TABLE vip_plans
    ADD COLUMN plan_type VARCHAR(16) NOT NULL DEFAULT 'normal' AFTER auto_renew_enabled;

UPDATE vip_plans
SET plan_type = CASE WHEN auto_renew_enabled = 1 THEN 'auto_renew' ELSE 'normal' END;

ALTER TABLE vip_plans
    ADD CONSTRAINT chk_vip_plans_type CHECK (plan_type IN ('normal', 'auto_renew'));

ALTER TABLE vip_agreements
    DROP INDEX uk_vip_agreements_key,
    ADD UNIQUE KEY uk_vip_agreements_key_version (agreement_key, version);

ALTER TABLE payment_orders
    ADD COLUMN service_agreement_id BIGINT UNSIGNED NULL AFTER plan_id,
    ADD COLUMN service_agreement_version INT UNSIGNED NULL AFTER service_agreement_id,
    ADD COLUMN auto_renew_agreement_id BIGINT UNSIGNED NULL AFTER service_agreement_version,
    ADD COLUMN auto_renew_agreement_version INT UNSIGNED NULL AFTER auto_renew_agreement_id,
    ADD KEY idx_payment_orders_service_agreement (service_agreement_id),
    ADD KEY idx_payment_orders_auto_renew_agreement (auto_renew_agreement_id),
    ADD CONSTRAINT fk_payment_orders_service_agreement FOREIGN KEY (service_agreement_id) REFERENCES vip_agreements (id) ON DELETE RESTRICT,
    ADD CONSTRAINT fk_payment_orders_auto_renew_agreement FOREIGN KEY (auto_renew_agreement_id) REFERENCES vip_agreements (id) ON DELETE RESTRICT;
