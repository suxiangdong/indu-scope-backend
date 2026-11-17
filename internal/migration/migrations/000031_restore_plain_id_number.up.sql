SET @id_number_column_exists := (
    SELECT COUNT(*)
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'personal_verifications'
      AND column_name = 'id_number'
);

SET @restore_id_number_sql := IF(
    @id_number_column_exists = 0,
    'ALTER TABLE personal_verifications ADD COLUMN id_number VARCHAR(32) NULL AFTER real_name',
    'SELECT 1'
);

PREPARE restore_id_number_statement FROM @restore_id_number_sql;
EXECUTE restore_id_number_statement;
DEALLOCATE PREPARE restore_id_number_statement;
