CREATE TABLE categories (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    parent_id BIGINT UNSIGNED NULL,
    name VARCHAR(100) NOT NULL,
    sort_order INT UNSIGNED NOT NULL DEFAULT 0,
    status TINYINT UNSIGNED NOT NULL DEFAULT 1,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    KEY idx_categories_parent_sort (parent_id, sort_order),
    CONSTRAINT chk_categories_status CHECK (status IN (0, 1)),
    CONSTRAINT fk_categories_parent FOREIGN KEY (parent_id) REFERENCES categories (id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE attributes (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    code VARCHAR(64) NOT NULL,
    name VARCHAR(100) NOT NULL,
    value_type TINYINT UNSIGNED NOT NULL,
    input_type VARCHAR(32) NOT NULL,
    unit VARCHAR(32) NULL,
    placeholder VARCHAR(255) NULL,
    validation_rules JSON NULL,
    help_text VARCHAR(255) NULL,
    status TINYINT UNSIGNED NOT NULL DEFAULT 1,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    UNIQUE KEY uk_attributes_code (code),
    CONSTRAINT chk_attributes_value_type CHECK (value_type IN (1, 2, 3, 4, 5, 6)),
    CONSTRAINT chk_attributes_status CHECK (status IN (0, 1))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE category_attributes (
    category_id BIGINT UNSIGNED NOT NULL,
    attribute_id BIGINT UNSIGNED NOT NULL,
    is_required TINYINT UNSIGNED NOT NULL DEFAULT 0,
    is_filterable TINYINT UNSIGNED NOT NULL DEFAULT 0,
    is_searchable TINYINT UNSIGNED NOT NULL DEFAULT 0,
    sort_order INT UNSIGNED NOT NULL DEFAULT 0,
    group_name VARCHAR(100) NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (category_id, attribute_id),
    KEY idx_category_attributes_attribute (attribute_id),
    CONSTRAINT fk_category_attributes_category FOREIGN KEY (category_id) REFERENCES categories (id) ON DELETE CASCADE,
    CONSTRAINT fk_category_attributes_attribute FOREIGN KEY (attribute_id) REFERENCES attributes (id) ON DELETE RESTRICT,
    CONSTRAINT chk_category_attributes_required CHECK (is_required IN (0, 1)),
    CONSTRAINT chk_category_attributes_filterable CHECK (is_filterable IN (0, 1)),
    CONSTRAINT chk_category_attributes_searchable CHECK (is_searchable IN (0, 1))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE attribute_options (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    attribute_id BIGINT UNSIGNED NOT NULL,
    code VARCHAR(64) NOT NULL,
    name VARCHAR(100) NOT NULL,
    sort_order INT UNSIGNED NOT NULL DEFAULT 0,
    status TINYINT UNSIGNED NOT NULL DEFAULT 1,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    UNIQUE KEY uk_attribute_options_attribute_code (attribute_id, code),
    KEY idx_attribute_options_attribute_sort (attribute_id, sort_order),
    CONSTRAINT fk_attribute_options_attribute FOREIGN KEY (attribute_id) REFERENCES attributes (id) ON DELETE CASCADE,
    CONSTRAINT chk_attribute_options_status CHECK (status IN (0, 1))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
