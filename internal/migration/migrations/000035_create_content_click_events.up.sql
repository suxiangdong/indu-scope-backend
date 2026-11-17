CREATE TABLE content_click_events (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    content_type VARCHAR(32) NOT NULL,
    content_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    KEY idx_content_click_events_type_time (content_type, created_at),
    KEY idx_content_click_events_content_time (content_type, content_id, created_at),
    KEY idx_content_click_events_user_time (user_id, created_at),
    CONSTRAINT fk_content_click_events_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE SET NULL,
    CONSTRAINT chk_content_click_events_type CHECK (content_type IN ('product', 'requirement'))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
