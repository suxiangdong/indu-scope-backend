ALTER TABLE procurement_requirements
    ADD COLUMN title VARCHAR(200) NULL AFTER stock_unit_name_snapshot;
