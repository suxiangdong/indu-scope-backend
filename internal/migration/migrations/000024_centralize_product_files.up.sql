ALTER TABLE product_images
    ADD COLUMN user_file_id BIGINT UNSIGNED NULL AFTER product_id;

-- Existing product images were uploaded into user_files before they were
-- linked from product_images. Recover that relationship before removing the
-- duplicated OSS metadata columns.
UPDATE product_images AS image
JOIN products AS product ON product.id = image.product_id
JOIN user_files AS file_record
  ON file_record.user_id = product.supplier_user_id
 AND file_record.purpose = 'product_image'
 AND file_record.bucket = image.bucket
 AND file_record.object_key = image.object_key
SET image.user_file_id = file_record.id
WHERE image.user_file_id IS NULL;

ALTER TABLE product_images
    MODIFY COLUMN user_file_id BIGINT UNSIGNED NOT NULL,
    DROP KEY uk_product_images_bucket_object,
    ADD UNIQUE KEY uk_product_images_product_file (product_id, user_file_id),
    ADD KEY idx_product_images_file (user_file_id),
    ADD CONSTRAINT fk_product_images_user_file FOREIGN KEY (user_file_id) REFERENCES user_files (id) ON DELETE RESTRICT,
    DROP COLUMN endpoint,
    DROP COLUMN bucket,
    DROP COLUMN object_key,
    DROP COLUMN original_name,
    DROP COLUMN content_type,
    DROP COLUMN file_size;
