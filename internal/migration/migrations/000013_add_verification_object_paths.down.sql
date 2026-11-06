ALTER TABLE company_verifications
    DROP COLUMN business_license_path,
    DROP COLUMN business_license_bucket;

ALTER TABLE personal_verifications
    DROP COLUMN id_back_path,
    DROP COLUMN id_back_bucket,
    DROP COLUMN id_front_path,
    DROP COLUMN id_front_bucket;
