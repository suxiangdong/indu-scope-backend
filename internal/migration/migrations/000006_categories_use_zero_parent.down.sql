ALTER TABLE categories
    MODIFY COLUMN parent_id BIGINT UNSIGNED NULL DEFAULT NULL;

UPDATE categories
SET parent_id = NULL
WHERE parent_id = 0;

ALTER TABLE categories
    ADD CONSTRAINT fk_categories_parent
    FOREIGN KEY (parent_id) REFERENCES categories (id) ON DELETE RESTRICT;
