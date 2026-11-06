ALTER TABLE user_files
    ADD COLUMN endpoint VARCHAR(255) NULL AFTER bucket;

UPDATE user_files
SET endpoint = 'oss-cn-beijing.aliyuncs.com'
WHERE endpoint IS NULL OR endpoint = '';

ALTER TABLE user_files
    MODIFY COLUMN endpoint VARCHAR(255) NOT NULL;

ALTER TABLE personal_verifications
    ADD COLUMN id_front_endpoint VARCHAR(255) NULL AFTER id_front_bucket,
    ADD COLUMN id_back_endpoint VARCHAR(255) NULL AFTER id_back_bucket;

ALTER TABLE company_verifications
    ADD COLUMN business_license_endpoint VARCHAR(255) NULL AFTER business_license_bucket;

UPDATE personal_verifications AS verification
JOIN user_files AS file_record
  ON file_record.user_id = verification.user_id
 AND file_record.bucket = verification.id_front_bucket
 AND file_record.object_key = verification.id_front_path
SET verification.id_front_endpoint = file_record.endpoint
WHERE verification.id_front_bucket IS NOT NULL
  AND verification.id_front_path IS NOT NULL;

UPDATE personal_verifications AS verification
JOIN user_files AS file_record
  ON file_record.user_id = verification.user_id
 AND file_record.bucket = verification.id_back_bucket
 AND file_record.object_key = verification.id_back_path
SET verification.id_back_endpoint = file_record.endpoint
WHERE verification.id_back_bucket IS NOT NULL
  AND verification.id_back_path IS NOT NULL;

UPDATE company_verifications AS verification
JOIN user_files AS file_record
  ON file_record.user_id = verification.user_id
 AND file_record.bucket = verification.business_license_bucket
 AND file_record.object_key = verification.business_license_path
SET verification.business_license_endpoint = file_record.endpoint
WHERE verification.business_license_bucket IS NOT NULL
  AND verification.business_license_path IS NOT NULL;
