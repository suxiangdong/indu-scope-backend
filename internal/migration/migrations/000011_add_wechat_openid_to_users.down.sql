ALTER TABLE users
    DROP INDEX uk_users_wechat_openid,
    DROP COLUMN wechat_openid;
