ALTER TABLE admin_users
    DROP CHECK chk_admin_users_super_admin,
    DROP COLUMN is_super_admin;
