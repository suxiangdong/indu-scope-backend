CREATE TABLE store_platforms (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,
    sort_order INT UNSIGNED NOT NULL DEFAULT 0,
    status TINYINT UNSIGNED NOT NULL DEFAULT 1,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    UNIQUE KEY uk_store_platforms_name (name),
    KEY idx_store_platforms_status_sort (status, sort_order),
    CONSTRAINT chk_store_platforms_status CHECK (status IN (0, 1))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO store_platforms (name, sort_order, status) VALUES
    ('京东', 10, 1),
    ('天猫', 20, 1),
    ('自建网站', 30, 1);

ALTER TABLE personal_verifications
    ADD COLUMN id_number VARCHAR(32) NULL AFTER real_name,
    ADD COLUMN store_platform_id BIGINT UNSIGNED NULL AFTER id_number,
    ADD COLUMN store_url VARCHAR(500) NULL AFTER store_platform_id,
    ADD KEY idx_personal_verifications_store_platform (store_platform_id),
    ADD CONSTRAINT fk_personal_verifications_store_platform
        FOREIGN KEY (store_platform_id) REFERENCES store_platforms (id) ON DELETE SET NULL;

-- id_number_encrypted is kept temporarily for compatibility with historical
-- rows. New and resubmitted records use the plaintext id_number column.
