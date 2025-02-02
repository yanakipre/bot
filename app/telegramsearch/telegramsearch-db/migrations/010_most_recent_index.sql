---- tern: disable-tx ----
SELECT 1;
-- TODO: uncomment
--
-- CREATE INDEX CONCURRENTLY IF NOT EXISTS
--     embeddings_most_recent_message_at_idx
--     ON embeddings (most_recent_message_at);
