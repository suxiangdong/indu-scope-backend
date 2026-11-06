ALTER TABLE personal_verifications
    ADD COLUMN review_status TINYINT UNSIGNED NOT NULL DEFAULT 0 AFTER verified_at,
    ADD COLUMN review_reason VARCHAR(500) NULL AFTER review_status,
    ADD COLUMN reviewed_by BIGINT UNSIGNED NULL AFTER review_reason,
    ADD COLUMN reviewed_at DATETIME(3) NULL AFTER reviewed_by,
    ADD KEY idx_personal_verifications_review_status (review_status),
    ADD KEY idx_personal_verifications_reviewer (reviewed_by),
    ADD CONSTRAINT chk_personal_verifications_review_status CHECK (review_status IN (0, 1, 2)),
    ADD CONSTRAINT fk_personal_verifications_reviewer FOREIGN KEY (reviewed_by) REFERENCES admin_users (id) ON DELETE SET NULL;

ALTER TABLE company_verifications
    ADD COLUMN review_status TINYINT UNSIGNED NOT NULL DEFAULT 0 AFTER verified_at,
    ADD COLUMN review_reason VARCHAR(500) NULL AFTER review_status,
    ADD COLUMN reviewed_by BIGINT UNSIGNED NULL AFTER review_reason,
    ADD COLUMN reviewed_at DATETIME(3) NULL AFTER reviewed_by,
    ADD KEY idx_company_verifications_review_status (review_status),
    ADD KEY idx_company_verifications_reviewer (reviewed_by),
    ADD CONSTRAINT chk_company_verifications_review_status CHECK (review_status IN (0, 1, 2)),
    ADD CONSTRAINT fk_company_verifications_reviewer FOREIGN KEY (reviewed_by) REFERENCES admin_users (id) ON DELETE SET NULL;

-- Existing records with a verification timestamp are already approved.
UPDATE personal_verifications
SET review_status = 1
WHERE verified_at IS NOT NULL;

UPDATE company_verifications
SET review_status = 1
WHERE verified_at IS NOT NULL;
