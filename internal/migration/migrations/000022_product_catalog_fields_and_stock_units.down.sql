ALTER TABLE products
    ADD COLUMN description_html LONGTEXT NULL AFTER name,
    ADD COLUMN condition_code VARCHAR(32) NOT NULL DEFAULT 'new' AFTER description_html,
    ADD COLUMN stock_unit_code VARCHAR(32) NOT NULL DEFAULT 'piece' AFTER stock_quantity;

UPDATE products AS product
JOIN stock_units AS stock_unit ON stock_unit.id = product.stock_unit_id
SET product.stock_unit_code = stock_unit.code;

DROP TABLE IF EXISTS product_images;

ALTER TABLE products
    DROP FOREIGN KEY fk_products_stock_unit,
    DROP KEY idx_products_stock_unit,
    DROP COLUMN stock_unit_id;

DROP TABLE IF EXISTS category_stock_units;
DROP TABLE IF EXISTS stock_units;
