DROP TRIGGER IF EXISTS outbox_notify_trigger ON outbox_events;
DROP FUNCTION IF EXISTS outbox_notify();
DROP INDEX IF EXISTS idx_outbox_unpublished;
DROP TABLE IF EXISTS outbox_events;
