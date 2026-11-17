-- 仅用于开发环境。
-- 初始登录账号：cffyadmin
-- 初始密码：123456
INSERT INTO admin_users (username, name, phone, password_hash, status, is_super_admin)
VALUES (
    'cffyadmin',
    '超级管理员',
    '10000000000',
    '$2a$10$oYagaUAft7tgfLr51D518OHurFl59FNGfGQdXuWGNnyIVaQi0HhDm',
    1,
    1
);
