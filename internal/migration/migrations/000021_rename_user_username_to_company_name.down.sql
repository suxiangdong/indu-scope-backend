ALTER TABLE users
    ADD COLUMN username VARCHAR(255) NULL AFTER phone;

UPDATE users
SET username = NULLIF(TRIM(company_name), '')
WHERE company_name IS NOT NULL
  AND TRIM(company_name) <> '';

ALTER TABLE users
    DROP COLUMN company_name;
