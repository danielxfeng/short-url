DROP INDEX IF EXISTS idx_links_active_by_user_created_at;

CREATE INDEX idx_links_by_user_id_desc
ON links (user_id, id DESC);
