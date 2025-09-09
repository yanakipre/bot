
ALTER TABLE chatthreads
    ADD COLUMN content_hash TEXT NOT NULL;

CREATE UNIQUE INDEX chatthreads_chat_id_content_hash_unique 
ON chatthreads (chat_id, content_hash);

---- create above / drop below ----


DROP INDEX IF EXISTS chatthreads_chat_id_content_hash_unique;

ALTER TABLE chatthreads
    DROP COLUMN IF EXISTS content_hash;
