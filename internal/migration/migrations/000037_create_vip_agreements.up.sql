CREATE TABLE vip_agreements (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    agreement_key VARCHAR(32) NOT NULL,
    title VARCHAR(100) NOT NULL,
    content TEXT NOT NULL,
    status TINYINT UNSIGNED NOT NULL DEFAULT 1,
    version INT UNSIGNED NOT NULL DEFAULT 1,
    updated_by BIGINT UNSIGNED NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    UNIQUE KEY uk_vip_agreements_key (agreement_key),
    KEY idx_vip_agreements_status (status),
    CONSTRAINT fk_vip_agreements_updater FOREIGN KEY (updated_by) REFERENCES admin_users (id) ON DELETE SET NULL,
    CONSTRAINT chk_vip_agreements_status CHECK (status IN (0, 1))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO vip_agreements (agreement_key, title, content, status, version)
VALUES
    ('service', '会员服务协议',
     '欢迎使用源仓会员服务。会员可享受平台提供的产品展示、采购需求响应、联系人服务等权益，具体权益以当前页面和平台实际配置为准。会员服务开通后，平台将在有效期内提供对应服务；如遇服务调整，平台会按照相关规则进行通知。', 1, 1),
    ('auto-renew', '自动续费协议',
     '如你选择自动续费，平台将在当前会员到期前按照所选套餐和页面展示的价格发起续费。你可以在会员中心随时取消自动续费，取消后当前会员权益仍可使用至已购买的服务周期结束。实际扣款以微信支付页面和相关支付规则为准。', 1, 1);
