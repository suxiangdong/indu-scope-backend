ALTER TABLE payment_orders
    DROP FOREIGN KEY fk_payment_orders_auto_renew_agreement,
    DROP FOREIGN KEY fk_payment_orders_service_agreement,
    DROP KEY idx_payment_orders_auto_renew_agreement,
    DROP KEY idx_payment_orders_service_agreement,
    DROP COLUMN auto_renew_agreement_version,
    DROP COLUMN auto_renew_agreement_id,
    DROP COLUMN service_agreement_version,
    DROP COLUMN service_agreement_id;

ALTER TABLE vip_agreements
    DROP INDEX uk_vip_agreements_key_version,
    ADD UNIQUE KEY uk_vip_agreements_key (agreement_key);

ALTER TABLE vip_plans
    DROP CHECK chk_vip_plans_type,
    DROP COLUMN plan_type;
