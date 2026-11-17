CREATE TABLE app_page_contents (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    page_key VARCHAR(32) NOT NULL,
    title VARCHAR(100) NOT NULL,
    summary VARCHAR(500) NULL,
    body TEXT NOT NULL,
    contact_phone VARCHAR(50) NULL,
    contact_email VARCHAR(120) NULL,
    service_hours VARCHAR(120) NULL,
    address VARCHAR(255) NULL,
    wechat_account VARCHAR(100) NULL,
    status TINYINT UNSIGNED NOT NULL DEFAULT 1,
    updated_by BIGINT UNSIGNED NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    UNIQUE KEY uk_app_page_contents_key (page_key),
    KEY idx_app_page_contents_status (status),
    CONSTRAINT fk_app_page_contents_updater FOREIGN KEY (updated_by) REFERENCES admin_users (id) ON DELETE SET NULL,
    CONSTRAINT chk_app_page_contents_status CHECK (status IN (0, 1))
);

INSERT INTO app_page_contents
    (page_key, title, summary, body, contact_phone, contact_email, service_hours, address, wechat_account, status)
VALUES
    ('about-us', '关于我们', '连接可信企业与工业品供需，让采购更简单。',
     '工业源Scope专注于工业品供需撮合，汇聚经过认证的企业与源头供应商，为采购方提供清晰、可靠、高效的产品信息与需求对接服务。\n\n我们坚持真实认证、源头直连和高效协同，帮助企业减少沟通成本，提升采购效率。',
     NULL, NULL, NULL, NULL, NULL, 1),
    ('contact', '联系客服', '如果你在使用过程中遇到问题，欢迎联系我们。',
     '我们的客服团队会协助处理登录、认证、产品、需求发布及平台使用相关问题。',
     '400-800-2024', 'service@indu-scope.com', '周一至周五 09:00-18:00', '北京市朝阳区工业品服务中心', 'indu-scope-service', 1);
