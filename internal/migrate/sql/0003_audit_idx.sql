-- Additional index to support the audit-log resource-id lookup hot path.
-- The composite idx_audit_resource (resource_type, resource_id) does not
-- serve a resource_id-only filter (leftmost-prefix rule), so a standalone
-- index on resource_id is useful.
--
-- MySQL does not support CREATE INDEX IF NOT EXISTS, so guard idempotency via
-- information_schema + a prepared statement. This makes the migration safe to
-- retry even though DDL auto-commits outside the runner's transaction.
SET @n := (
  SELECT COUNT(*) FROM information_schema.statistics
  WHERE table_schema = DATABASE()
    AND table_name = 'audit_logs'
    AND index_name = 'idx_audit_resource_id'
);
SET @s := IF(@n = 0,
  'CREATE INDEX idx_audit_resource_id ON audit_logs (resource_id)',
  'SELECT 1');
PREPARE stmt FROM @s;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

INSERT INTO schema_migrations (version) VALUES (3)
ON DUPLICATE KEY UPDATE version = version;
