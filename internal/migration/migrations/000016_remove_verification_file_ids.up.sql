ALTER TABLE personal_verifications
    DROP COLUMN id_front_file_id,
    DROP COLUMN id_back_file_id;

ALTER TABLE company_verifications
    DROP COLUMN business_license_file_id;
