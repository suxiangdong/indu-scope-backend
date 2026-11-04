-- Store 0 for top-level categories instead of NULL. A zero parent does not
-- refer to another row, so the old self-referencing foreign key is removed.
ALTER TABLE categories
    DROP FOREIGN KEY fk_categories_parent;

UPDATE categories
SET parent_id = 0
WHERE parent_id IS NULL;

ALTER TABLE categories
    MODIFY COLUMN parent_id BIGINT UNSIGNED NOT NULL DEFAULT 0;
