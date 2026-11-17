ALTER TABLE company_profiles
    ADD COLUMN store_platform_id BIGINT UNSIGNED NULL AFTER introduction_html,
    ADD COLUMN store_url VARCHAR(500) NULL AFTER store_platform_id,
    ADD KEY idx_company_profiles_store_platform (store_platform_id),
    ADD CONSTRAINT fk_company_profiles_store_platform
        FOREIGN KEY (store_platform_id) REFERENCES store_platforms (id) ON DELETE SET NULL;
