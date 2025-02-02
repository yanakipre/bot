---- tern: disable-tx ----

CREATE INDEX CONCURRENTLY IF NOT EXISTS
    embeddings_most_recent_message_at_idx
    ON chatthreads (most_recent_message_at);
