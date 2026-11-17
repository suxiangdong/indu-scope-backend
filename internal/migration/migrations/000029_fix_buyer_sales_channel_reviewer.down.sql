UPDATE buyer_sales_channels AS channel
LEFT JOIN users AS user_record ON user_record.id = channel.reviewed_by
SET channel.reviewed_by = NULL
WHERE channel.reviewed_by IS NOT NULL
  AND user_record.id IS NULL;

ALTER TABLE buyer_sales_channels
    DROP FOREIGN KEY fk_buyer_sales_channels_reviewer;

ALTER TABLE buyer_sales_channels
    ADD CONSTRAINT fk_buyer_sales_channels_reviewer
        FOREIGN KEY (reviewed_by) REFERENCES users (id) ON DELETE SET NULL;
