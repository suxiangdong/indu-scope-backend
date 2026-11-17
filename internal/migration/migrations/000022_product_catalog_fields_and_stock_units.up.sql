CREATE TABLE stock_units (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    code VARCHAR(32) NOT NULL,
    name VARCHAR(32) NOT NULL,
    sort_order INT UNSIGNED NOT NULL DEFAULT 0,
    status TINYINT UNSIGNED NOT NULL DEFAULT 1,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    UNIQUE KEY uk_stock_units_code (code),
    UNIQUE KEY uk_stock_units_name (name),
    KEY idx_stock_units_status_sort (status, sort_order),
    CONSTRAINT chk_stock_units_status CHECK (status IN (0, 1))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO stock_units (code, name, sort_order, status)
VALUES ('piece', '件', 0, 1);

-- Preserve the existing free-text stock unit values before replacing the
-- products.stock_unit_code column. The fallback to "件" below only applies
-- when a historical product did not have a meaningful unit.
INSERT IGNORE INTO stock_units (code, name, sort_order, status)
SELECT DISTINCT TRIM(stock_unit_code), TRIM(stock_unit_code), 100, 1
FROM products
WHERE NULLIF(TRIM(stock_unit_code), '') IS NOT NULL
  AND TRIM(stock_unit_code) <> 'piece';

CREATE TABLE category_stock_units (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    category_id BIGINT UNSIGNED NOT NULL,
    stock_unit_id BIGINT UNSIGNED NOT NULL,
    sort_order INT UNSIGNED NOT NULL DEFAULT 0,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    UNIQUE KEY uk_category_stock_units_category_unit (category_id, stock_unit_id),
    KEY idx_category_stock_units_unit (stock_unit_id),
    CONSTRAINT fk_category_stock_units_category FOREIGN KEY (category_id) REFERENCES categories (id) ON DELETE CASCADE,
    CONSTRAINT fk_category_stock_units_unit FOREIGN KEY (stock_unit_id) REFERENCES stock_units (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Keep existing products publishable after the free-text unit is replaced by a
-- managed unit. Existing values are normalized to the seeded "件" unit.
ALTER TABLE products
    ADD COLUMN stock_unit_id BIGINT UNSIGNED NULL AFTER name;

UPDATE products
SET stock_unit_id = (SELECT id FROM stock_units WHERE code = 'piece' LIMIT 1)
WHERE stock_unit_id IS NULL;

UPDATE products AS product
JOIN stock_units AS stock_unit ON stock_unit.code = NULLIF(TRIM(product.stock_unit_code), '')
SET product.stock_unit_id = stock_unit.id;

ALTER TABLE products
    MODIFY COLUMN stock_unit_id BIGINT UNSIGNED NOT NULL,
    ADD KEY idx_products_stock_unit (stock_unit_id),
    ADD CONSTRAINT fk_products_stock_unit FOREIGN KEY (stock_unit_id) REFERENCES stock_units (id) ON DELETE RESTRICT,
    DROP COLUMN description_html,
    DROP COLUMN condition_code,
    DROP COLUMN stock_unit_code;

-- Bind the default unit to existing leaf categories. Administrators can
-- replace these bindings later; categories without bindings continue to use
-- all active units as a fallback.
INSERT INTO category_stock_units (category_id, stock_unit_id, sort_order)
SELECT category.id, stock_unit.id, 0
FROM categories AS category
JOIN stock_units AS stock_unit ON stock_unit.code = 'piece'
WHERE NOT EXISTS (
    SELECT 1
    FROM categories AS child
    WHERE child.parent_id = category.id
);

CREATE TABLE product_images (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    product_id BIGINT UNSIGNED NOT NULL,
    endpoint VARCHAR(255) NOT NULL,
    bucket VARCHAR(128) NOT NULL,
    object_key VARCHAR(512) NOT NULL,
    original_name VARCHAR(255) NOT NULL,
    content_type VARCHAR(128) NOT NULL,
    file_size BIGINT UNSIGNED NOT NULL,
    sort_order INT UNSIGNED NOT NULL DEFAULT 0,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    UNIQUE KEY uk_product_images_product_order (product_id, sort_order),
    UNIQUE KEY uk_product_images_bucket_object (bucket, object_key),
    KEY idx_product_images_product (product_id),
    CONSTRAINT fk_product_images_product FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
