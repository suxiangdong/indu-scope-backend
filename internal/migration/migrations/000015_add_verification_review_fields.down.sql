ALTER TABLE company_verifications
    DROP FOREIGN KEY fk_company_verifications_reviewer,
    DROP CHECK chk_company_verifications_review_status,
    DROP KEY idx_company_verifications_reviewer,
    DROP KEY idx_company_verifications_review_status,
    DROP COLUMN reviewed_at,
    DROP COLUMN reviewed_by,
    DROP COLUMN review_reason,
    DROP COLUMN review_status;

ALTER TABLE personal_verifications
    DROP FOREIGN KEY fk_personal_verifications_reviewer,
    DROP CHECK chk_personal_verifications_review_status,
    DROP KEY idx_personal_verifications_reviewer,
    DROP KEY idx_personal_verifications_review_status,
    DROP COLUMN reviewed_at,
    DROP COLUMN reviewed_by,
    DROP COLUMN review_reason,
    DROP COLUMN review_status;
