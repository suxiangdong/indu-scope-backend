ALTER TABLE users
    ADD COLUMN username VARCHAR(255) NULL AFTER phone;

UPDATE users u
JOIN company_verifications cv ON cv.user_id = u.id
SET u.username = NULLIF(TRIM(cv.company_name), '')
WHERE cv.review_status = 1
  AND TRIM(cv.company_name) <> '';
