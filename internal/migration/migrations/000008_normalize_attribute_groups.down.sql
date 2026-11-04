ALTER TABLE category_attributes
    DROP FOREIGN KEY fk_category_attributes_group;

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

INSERT INTO category_attribute_groups (category_id, code, name, sort_order, status)
SELECT
    binding.category_id,
    CONCAT('group_', MD5(CONCAT(CAST(binding.category_id AS CHAR), ':', group_table.name))),
    group_table.name,
    binding.sort_order,
    binding.status
FROM category_attribute_group_bindings binding
JOIN attribute_groups group_table ON group_table.id = binding.group_id;

UPDATE category_attributes ca
JOIN category_attribute_group_bindings binding
  ON binding.category_id = ca.category_id
 AND binding.group_id = ca.group_id
JOIN category_attribute_groups old_group
  ON old_group.category_id = binding.category_id
 AND old_group.name = (
     SELECT name FROM attribute_groups WHERE id = binding.group_id
 )
SET ca.group_id = old_group.id,
    ca.group_name = old_group.name;

ALTER TABLE category_attributes
    ADD CONSTRAINT fk_category_attributes_group
        FOREIGN KEY (group_id) REFERENCES category_attribute_groups (id) ON DELETE SET NULL;

DROP TABLE category_attribute_group_bindings;
DROP TABLE attribute_group_attributes;
DROP TABLE attribute_groups;
