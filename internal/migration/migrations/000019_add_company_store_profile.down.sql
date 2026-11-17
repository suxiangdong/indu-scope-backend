ALTER TABLE company_profiles
    DROP FOREIGN KEY fk_company_profiles_store_platform,
    DROP KEY idx_company_profiles_store_platform,
    DROP COLUMN store_url,
    DROP COLUMN store_platform_id;
