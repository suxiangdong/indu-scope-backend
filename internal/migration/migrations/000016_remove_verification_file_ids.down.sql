ALTER TABLE personal_verifications
    ADD COLUMN id_front_file_id VARCHAR(255) NULL AFTER id_number_encrypted,
    ADD COLUMN id_back_file_id VARCHAR(255) NULL AFTER id_front_path;

ALTER TABLE company_verifications
    ADD COLUMN business_license_file_id VARCHAR(255) NULL AFTER unified_social_credit_code;
