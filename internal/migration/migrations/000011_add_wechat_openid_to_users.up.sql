ALTER TABLE users
    ADD COLUMN wechat_openid VARCHAR(64) NULL AFTER phone,
    ADD UNIQUE KEY uk_users_wechat_openid (wechat_openid);
