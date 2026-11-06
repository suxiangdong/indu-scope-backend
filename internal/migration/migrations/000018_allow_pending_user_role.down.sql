-- A rollback cannot preserve the pending state because the old schema did not
-- allow role=0. Resolve pending users as buyers before restoring the old check.
UPDATE users SET role = 1 WHERE role = 0;

ALTER TABLE users
    DROP CHECK chk_users_role,
    ADD CONSTRAINT chk_users_role CHECK (role IN (1, 2));
