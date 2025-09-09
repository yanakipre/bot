
ALTER TABLE chatthreads
    ADD COLUMN thread_starter_id BIGINT;

DROP INDEX IF EXISTS chatthreads_chat_id_content_hash_unique;

CREATE UNIQUE INDEX chatthreads_chat_id_thread_starter_id_unique 
ON chatthreads (chat_id, thread_starter_id);

ALTER TABLE chatthreads
    DROP COLUMN IF EXISTS content_hash;

ALTER TABLE chatmessages DROP CONSTRAINT IF EXISTS chatmessages_pkey;

ALTER TABLE chatmessages ADD CONSTRAINT chatmessages_pkey PRIMARY KEY (chat_id, message_id);

---- create above / drop below ----


DROP INDEX IF EXISTS chatthreads_chat_id_thread_starter_id_unique;

CREATE UNIQUE INDEX chatthreads_chat_id_content_hash_unique 
ON chatthreads (chat_id, content_hash);

ALTER TABLE chatthreads
    DROP COLUMN IF EXISTS thread_starter_id;
