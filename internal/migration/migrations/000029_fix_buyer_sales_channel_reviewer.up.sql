UPDATE buyer_sales_channels AS channel
LEFT JOIN admin_users AS admin_record ON admin_record.id = channel.reviewed_by
SET channel.reviewed_by = NULL
WHERE channel.reviewed_by IS NOT NULL
  AND admin_record.id IS NULL;

ALTER TABLE buyer_sales_channels
    DROP FOREIGN KEY fk_buyer_sales_channels_reviewer;

ALTER TABLE buyer_sales_channels
    ADD CONSTRAINT fk_buyer_sales_channels_reviewer
        FOREIGN KEY (reviewed_by) REFERENCES admin_users (id) ON DELETE SET NULL;
