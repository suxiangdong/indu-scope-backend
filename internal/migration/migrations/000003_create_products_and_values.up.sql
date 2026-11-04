CREATE TABLE products (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    supplier_user_id BIGINT UNSIGNED NOT NULL,
    category_id BIGINT UNSIGNED NOT NULL,
    name VARCHAR(255) NOT NULL,
    description_html LONGTEXT NULL,
    condition_code VARCHAR(32) NOT NULL,
    stock_quantity DECIMAL(18, 3) NOT NULL DEFAULT 0,
    stock_unit_code VARCHAR(32) NOT NULL DEFAULT 'piece',
    stock_updated_at DATETIME(3) NULL,
    stock_updated_by BIGINT UNSIGNED NULL,
    status TINYINT UNSIGNED NOT NULL DEFAULT 0,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    KEY idx_products_supplier_status (supplier_user_id, status),
    KEY idx_products_category_status (category_id, status),
    CONSTRAINT chk_products_stock_quantity CHECK (stock_quantity >= 0),
    CONSTRAINT chk_products_status CHECK (status IN (0, 1, 2)),
    CONSTRAINT fk_products_supplier FOREIGN KEY (supplier_user_id) REFERENCES users (id) ON DELETE RESTRICT,
    CONSTRAINT fk_products_category FOREIGN KEY (category_id) REFERENCES categories (id) ON DELETE RESTRICT,
    CONSTRAINT fk_products_stock_updater FOREIGN KEY (stock_updated_by) REFERENCES users (id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE product_attribute_values (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    product_id BIGINT UNSIGNED NOT NULL,
    attribute_id BIGINT UNSIGNED NOT NULL,
    value_number DECIMAL(18, 6) NULL,
    value_text VARCHAR(1000) NULL,
    value_bool TINYINT UNSIGNED NULL,
    value_option_id BIGINT UNSIGNED NULL,
    value_json JSON NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    UNIQUE KEY uk_product_attribute_values_product_attribute (product_id, attribute_id),
    KEY idx_product_attribute_values_attribute_number (attribute_id, value_number, product_id),
    KEY idx_product_attribute_values_attribute_option (attribute_id, value_option_id, product_id),
    KEY idx_product_attribute_values_attribute_text (attribute_id, value_text(100), product_id),
    CONSTRAINT fk_product_attribute_values_product FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE CASCADE,
    CONSTRAINT fk_product_attribute_values_attribute FOREIGN KEY (attribute_id) REFERENCES attributes (id) ON DELETE RESTRICT,
    CONSTRAINT fk_product_attribute_values_option FOREIGN KEY (value_option_id) REFERENCES attribute_options (id) ON DELETE RESTRICT,
    CONSTRAINT chk_product_attribute_values_bool CHECK (value_bool IS NULL OR value_bool IN (0, 1))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE product_attribute_value_options (
    value_id BIGINT UNSIGNED NOT NULL,
    option_id BIGINT UNSIGNED NOT NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (value_id, option_id),
    KEY idx_product_attribute_value_options_option (option_id),
    CONSTRAINT fk_product_attribute_value_options_value FOREIGN KEY (value_id) REFERENCES product_attribute_values (id) ON DELETE CASCADE,
    CONSTRAINT fk_product_attribute_value_options_option FOREIGN KEY (option_id) REFERENCES attribute_options (id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
