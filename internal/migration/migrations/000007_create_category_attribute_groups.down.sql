ALTER TABLE category_attributes
    DROP FOREIGN KEY fk_category_attributes_group,
    DROP INDEX idx_category_attributes_group,
    DROP COLUMN group_id;

DROP TABLE IF EXISTS category_attribute_groups;
