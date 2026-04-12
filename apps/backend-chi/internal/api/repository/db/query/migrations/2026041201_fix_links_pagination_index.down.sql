DROP INDEX IF EXISTS idx_links_by_user_id_desc;

CREATE INDEX idx_links_active_by_user_created_at
ON links (user_id, created_at DESC)
WHERE deleted_at IS NULL;
