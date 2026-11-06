ALTER TABLE personal_verifications
    ADD COLUMN id_front_bucket VARCHAR(128) NULL AFTER id_front_file_id,
    ADD COLUMN id_front_path VARCHAR(512) NULL AFTER id_front_bucket,
    ADD COLUMN id_back_bucket VARCHAR(128) NULL AFTER id_back_file_id,
    ADD COLUMN id_back_path VARCHAR(512) NULL AFTER id_back_bucket;

ALTER TABLE company_verifications
    ADD COLUMN business_license_bucket VARCHAR(128) NULL AFTER business_license_file_id,
    ADD COLUMN business_license_path VARCHAR(512) NULL AFTER business_license_bucket;

-- Existing uploads created by the previous file-id flow are backfilled when
-- they can be resolved to the authenticated user's user_files record.
UPDATE company_verifications AS verification
JOIN user_files AS file_record
  ON file_record.id = CAST(verification.business_license_file_id AS UNSIGNED)
 AND file_record.user_id = verification.user_id
 AND file_record.purpose = 'business_license'
SET verification.business_license_bucket = file_record.bucket,
    verification.business_license_path = file_record.object_key
WHERE verification.business_license_file_id IS NOT NULL
  AND verification.business_license_file_id <> '';
