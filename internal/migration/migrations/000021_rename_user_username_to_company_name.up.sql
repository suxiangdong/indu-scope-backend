ALTER TABLE users
    ADD COLUMN company_name VARCHAR(255) NULL AFTER phone;

UPDATE users
SET company_name = NULLIF(TRIM(username), '')
WHERE username IS NOT NULL
  AND TRIM(username) <> '';

UPDATE users u
JOIN company_verifications cv ON cv.user_id = u.id
SET u.company_name = NULLIF(TRIM(cv.company_name), '')
WHERE cv.review_status = 1
  AND TRIM(cv.company_name) <> '';

ALTER TABLE users
    DROP COLUMN username;
