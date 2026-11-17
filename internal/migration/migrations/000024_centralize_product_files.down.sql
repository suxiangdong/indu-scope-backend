ALTER TABLE product_images
    ADD COLUMN endpoint VARCHAR(255) NULL AFTER product_id,
    ADD COLUMN bucket VARCHAR(128) NULL AFTER endpoint,
    ADD COLUMN object_key VARCHAR(512) NULL AFTER bucket,
    ADD COLUMN original_name VARCHAR(255) NULL AFTER object_key,
    ADD COLUMN content_type VARCHAR(128) NULL AFTER original_name,
    ADD COLUMN file_size BIGINT UNSIGNED NULL AFTER content_type;

UPDATE product_images AS image
JOIN user_files AS file_record ON file_record.id = image.user_file_id
SET image.endpoint = file_record.endpoint,
    image.bucket = file_record.bucket,
    image.object_key = file_record.object_key,
    image.original_name = file_record.original_name,
    image.content_type = file_record.content_type,
    image.file_size = file_record.file_size;

ALTER TABLE product_images
    MODIFY COLUMN endpoint VARCHAR(255) NOT NULL,
    MODIFY COLUMN bucket VARCHAR(128) NOT NULL,
    MODIFY COLUMN object_key VARCHAR(512) NOT NULL,
    MODIFY COLUMN original_name VARCHAR(255) NOT NULL,
    MODIFY COLUMN content_type VARCHAR(128) NOT NULL,
    MODIFY COLUMN file_size BIGINT UNSIGNED NOT NULL,
    DROP FOREIGN KEY fk_product_images_user_file,
    DROP KEY uk_product_images_product_file,
    DROP KEY idx_product_images_file,
    ADD UNIQUE KEY uk_product_images_bucket_object (bucket, object_key),
    DROP COLUMN user_file_id;
