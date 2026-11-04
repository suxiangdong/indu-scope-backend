CREATE TABLE category_attribute_groups (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    category_id BIGINT UNSIGNED NOT NULL,
    code VARCHAR(64) NOT NULL,
    name VARCHAR(100) NOT NULL,
    sort_order INT UNSIGNED NOT NULL DEFAULT 0,
    status TINYINT UNSIGNED NOT NULL DEFAULT 1,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    UNIQUE KEY uk_category_attribute_groups_category_code (category_id, code),
    UNIQUE KEY uk_category_attribute_groups_category_name (category_id, name),
    KEY idx_category_attribute_groups_category_sort (category_id, sort_order),
    CONSTRAINT chk_category_attribute_groups_status CHECK (status IN (0, 1)),
    CONSTRAINT fk_category_attribute_groups_category
        FOREIGN KEY (category_id) REFERENCES categories (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Convert existing free-text group names into stable, category-scoped groups.
INSERT INTO category_attribute_groups (category_id, code, name, sort_order)
SELECT
    category_id,
    CONCAT('group_', MD5(CONCAT(CAST(category_id AS CHAR), ':', TRIM(group_name)))),
    TRIM(group_name),
    MIN(sort_order)
FROM category_attributes
WHERE group_name IS NOT NULL AND TRIM(group_name) <> ''
GROUP BY category_id, TRIM(group_name);

ALTER TABLE category_attributes
    ADD COLUMN group_id BIGINT UNSIGNED NULL AFTER attribute_id,
    ADD KEY idx_category_attributes_group (group_id);

UPDATE category_attributes ca
JOIN category_attribute_groups g
  ON g.category_id = ca.category_id
 AND g.name = TRIM(ca.group_name)
SET ca.group_id = g.id
WHERE ca.group_name IS NOT NULL AND TRIM(ca.group_name) <> '';

ALTER TABLE category_attributes
    ADD CONSTRAINT fk_category_attributes_group
        FOREIGN KEY (group_id) REFERENCES category_attribute_groups (id) ON DELETE SET NULL;
