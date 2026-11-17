ALTER TABLE admin_users
    ADD COLUMN is_super_admin TINYINT UNSIGNED NOT NULL DEFAULT 0 AFTER status,
    ADD CONSTRAINT chk_admin_users_super_admin CHECK (is_super_admin IN (0, 1));

-- The development administrator is the initial super administrator. Production
-- deployments should explicitly mark their own administrator account instead.
UPDATE admin_users
SET is_super_admin = 1
WHERE username = 'cffyadmin';
