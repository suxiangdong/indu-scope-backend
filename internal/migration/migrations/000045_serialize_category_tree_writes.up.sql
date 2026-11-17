CREATE TABLE backend_write_locks (
    lock_key VARCHAR(64) NOT NULL,
    PRIMARY KEY (lock_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Structural category writes take this row FOR UPDATE within their transaction.
-- A persistent sentinel also protects empty trees and multiple app instances.
INSERT INTO backend_write_locks (lock_key) VALUES ('category_tree');
