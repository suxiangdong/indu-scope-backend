-- Give existing approved requirements the same finite lifetime as new ones.
UPDATE procurement_requirements
SET expires_at = DATE_ADD(COALESCE(published_at, created_at), INTERVAL 30 DAY)
WHERE review_status = 1
  AND status = 1
  AND expires_at IS NULL;
