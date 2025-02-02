CREATE EXTENSION vector;

CREATE TABLE embeddings
(
    thread_id BIGSERIAL NOT NULL,
    chat_id       TEXT      NOT NULL,
    message       TEXT      NOT NULL,
    embedding     VECTOR(2000),
    embedding_id  BIGSERIAL PRIMARY KEY
);

ALTER TABLE embeddings
    ADD CONSTRAINT embeddings_chat_id_fk FOREIGN KEY (chat_id) REFERENCES chats (chat_id) ON DELETE CASCADE;

CREATE INDEX embeddings_chat_id_idx ON embeddings USING hash (chat_id);

ALTER TABLE embeddings
    ADD CONSTRAINT embeddings_chatthread_id_fk FOREIGN KEY (thread_id) REFERENCES chatthreads (thread_id) ON DELETE CASCADE;

CREATE UNIQUE INDEX embeddings_chatthread_id_idx ON embeddings USING btree (thread_id);

---- create above / drop below ----

DROP TABLE embeddings;

DROP EXTENSION vector;
