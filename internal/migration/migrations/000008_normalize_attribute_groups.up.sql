-- Normalize attribute groups so they are reusable across categories.
CREATE TABLE attribute_groups (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    code VARCHAR(64) NOT NULL,
    name VARCHAR(100) NOT NULL,
    sort_order INT UNSIGNED NOT NULL DEFAULT 0,
    status TINYINT UNSIGNED NOT NULL DEFAULT 1,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    UNIQUE KEY uk_attribute_groups_code (code),
    UNIQUE KEY uk_attribute_groups_name (name),
    KEY idx_attribute_groups_sort (sort_order, id),
    CONSTRAINT chk_attribute_groups_status CHECK (status IN (0, 1))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE attribute_group_attributes (
    group_id BIGINT UNSIGNED NOT NULL,
    attribute_id BIGINT UNSIGNED NOT NULL,
    sort_order INT UNSIGNED NOT NULL DEFAULT 0,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (group_id, attribute_id),
    KEY idx_attribute_group_attributes_attribute_sort (attribute_id, sort_order),
    CONSTRAINT fk_attribute_group_attributes_group
        FOREIGN KEY (group_id) REFERENCES attribute_groups (id) ON DELETE CASCADE,
    CONSTRAINT fk_attribute_group_attributes_attribute
        FOREIGN KEY (attribute_id) REFERENCES attributes (id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE category_attribute_group_bindings (
    category_id BIGINT UNSIGNED NOT NULL,
    group_id BIGINT UNSIGNED NOT NULL,
    sort_order INT UNSIGNED NOT NULL DEFAULT 0,
    status TINYINT UNSIGNED NOT NULL DEFAULT 1,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (category_id, group_id),
    KEY idx_category_attribute_group_bindings_category_sort (category_id, sort_order),
    CONSTRAINT fk_category_attribute_group_bindings_category
        FOREIGN KEY (category_id) REFERENCES categories (id) ON DELETE CASCADE,
    CONSTRAINT fk_category_attribute_group_bindings_group
        FOREIGN KEY (group_id) REFERENCES attribute_groups (id) ON DELETE RESTRICT,
    CONSTRAINT chk_category_attribute_group_bindings_status CHECK (status IN (0, 1))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- The 000007 table stored category-scoped groups. Merge those groups by
-- name into the reusable global group table. Existing free-text group names
-- are included as a safety net for databases that had legacy rows.
INSERT IGNORE INTO attribute_groups (code, name, sort_order, status)
SELECT
    CONCAT('group_', MD5(TRIM(name))),
    TRIM(name),
    MIN(sort_order),
    MAX(status)
FROM category_attribute_groups
WHERE name IS NOT NULL AND TRIM(name) <> ''
GROUP BY TRIM(name);

INSERT IGNORE INTO attribute_groups (code, name, sort_order, status)
SELECT
    CONCAT('group_', MD5(TRIM(group_name))),
    TRIM(group_name),
    MIN(sort_order),
    1
FROM category_attributes
WHERE group_name IS NOT NULL AND TRIM(group_name) <> ''
GROUP BY TRIM(group_name);

-- Remove the old foreign key before changing group_id from the old table's
-- id space to the new global attribute_groups id space.
ALTER TABLE category_attributes
    DROP FOREIGN KEY fk_category_attributes_group;

UPDATE category_attributes ca
JOIN category_attribute_groups old_group ON old_group.id = ca.group_id
JOIN attribute_groups new_group ON new_group.name = TRIM(old_group.name)
SET ca.group_id = new_group.id,
    ca.group_name = new_group.name
WHERE ca.group_id IS NOT NULL;

UPDATE category_attributes ca
JOIN attribute_groups new_group ON new_group.name = TRIM(ca.group_name)
SET ca.group_id = new_group.id,
    ca.group_name = new_group.name
WHERE ca.group_name IS NOT NULL
  AND TRIM(ca.group_name) <> '';

-- Preserve old direct category bindings that had no group by putting them in
-- a shared system group. New attributes may still remain ungrouped until an
-- administrator assigns them to a reusable group.
INSERT IGNORE INTO attribute_groups (code, name, sort_order, status)
SELECT '__ungrouped', '未分组', 4294967295, 1
FROM DUAL
WHERE EXISTS (
    SELECT 1
    FROM category_attributes
    WHERE group_id IS NULL
      AND (group_name IS NULL OR TRIM(group_name) = '')
);

UPDATE category_attributes ca
JOIN attribute_groups new_group ON new_group.code = '__ungrouped'
SET ca.group_id = new_group.id,
    ca.group_name = new_group.name
WHERE ca.group_id IS NULL
  AND (ca.group_name IS NULL OR TRIM(ca.group_name) = '');

INSERT IGNORE INTO attribute_group_attributes (group_id, attribute_id, sort_order)
SELECT group_id, attribute_id, MIN(sort_order)
FROM category_attributes
WHERE group_id IS NOT NULL
GROUP BY group_id, attribute_id;

INSERT IGNORE INTO category_attribute_group_bindings (category_id, group_id, sort_order, status)
SELECT
    old_group.category_id,
    new_group.id,
    old_group.sort_order,
    old_group.status
FROM category_attribute_groups old_group
JOIN attribute_groups new_group ON new_group.name = TRIM(old_group.name);

INSERT IGNORE INTO category_attribute_group_bindings (category_id, group_id, sort_order, status)
SELECT category_id, group_id, MIN(sort_order), 1
FROM category_attributes
WHERE group_id IS NOT NULL
GROUP BY category_id, group_id;

ALTER TABLE category_attributes
    ADD CONSTRAINT fk_category_attributes_group
        FOREIGN KEY (group_id) REFERENCES attribute_groups (id) ON DELETE RESTRICT;

DROP TABLE category_attribute_groups;
