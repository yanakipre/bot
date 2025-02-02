CREATE TABLE chats (
    chat_id TEXT PRIMARY KEY
);

CREATE TABLE chatthreads (
    thread_id BIGSERIAL PRIMARY KEY,
    chat_id TEXT NOT NULL,
    body jsonb NOT NULL
);

ALTER TABLE chatthreads
    ADD CONSTRAINT chatthreads_chat_id_fk FOREIGN KEY (chat_id) REFERENCES chats(chat_id) ON DELETE CASCADE;

CREATE INDEX chatthreads_chat_id_idx ON chatthreads USING hash (chat_id);

---- create above / drop below ----

DROP TABLE chatthreads;

DROP TABLE chats;
