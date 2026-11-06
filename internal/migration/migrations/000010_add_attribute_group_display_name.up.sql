-- Keep the internal group name available for uniqueness and matching while
-- exposing a separate label that can be changed for frontend presentation.
ALTER TABLE attribute_groups
    ADD COLUMN display_name VARCHAR(100) NULL AFTER name;

UPDATE attribute_groups
SET display_name = name
WHERE display_name IS NULL;

ALTER TABLE attribute_groups
    MODIFY COLUMN display_name VARCHAR(100) NOT NULL;
