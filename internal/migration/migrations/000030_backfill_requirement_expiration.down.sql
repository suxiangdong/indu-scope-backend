UPDATE procurement_requirements
SET expires_at = NULL
WHERE review_status = 1
  AND status IN (1, 4)
  AND expires_at = DATE_ADD(COALESCE(published_at, created_at), INTERVAL 30 DAY);
