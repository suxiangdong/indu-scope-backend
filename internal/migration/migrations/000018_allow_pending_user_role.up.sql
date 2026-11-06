-- Newly registered frontend users do not have a business role until they
-- complete their profile. Keep 1/2 for buyer/supplier and use 0 for pending.
ALTER TABLE users
    DROP CHECK chk_users_role,
    ADD CONSTRAINT chk_users_role CHECK (role IN (0, 1, 2));
