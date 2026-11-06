ALTER TABLE personal_verifications
    DROP FOREIGN KEY fk_personal_verifications_store_platform,
    DROP KEY idx_personal_verifications_store_platform,
    DROP COLUMN store_url,
    DROP COLUMN store_platform_id,
    DROP COLUMN id_number;

DROP TABLE IF EXISTS store_platforms;
