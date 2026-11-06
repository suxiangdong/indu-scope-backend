ALTER TABLE company_verifications
    DROP COLUMN business_license_endpoint;

ALTER TABLE personal_verifications
    DROP COLUMN id_back_endpoint,
    DROP COLUMN id_front_endpoint;

ALTER TABLE user_files
    DROP COLUMN endpoint;
