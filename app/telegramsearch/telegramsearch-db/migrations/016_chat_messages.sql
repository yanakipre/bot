
CREATE TABLE chatmessages (
    message_id BIGINT NOT NULL PRIMARY KEY,
    chat_id TEXT NOT NULL,
    from_id bigint NOT NULL,
    text_entities TEXT NOT NULL, -- 
    message TEXT NOT NULL,
    reply_to bigint NULL,
    date TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT '2011-05-19T09:45:17',
    type TEXT NOT NULL 
);

ALTER TABLE chatmessages
    ADD CONSTRAINT chatmessages_chat_id_fk FOREIGN KEY (chat_id) REFERENCES chats(chat_id) ON DELETE CASCADE;

CREATE INDEX chatmessages_chat_id_idx ON chatmessages USING hash (chat_id);

ALTER TABLE chats
    ADD COLUMN last_read_id bigint NOT NULL DEFAULT 0;

ALTER TABLE chatthreads
    ADD COLUMN fresh_embeddings_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT '2011-05-19T09:45:17';


---- create above / drop below ----

ALTER TABLE chatthreads
    DROP COLUMN fresh_embeddings_at;

ALTER TABLE chats
    DROP COLUMN last_read_id;

DROP INDEX chatmessages_chat_id_idx;

ALTER TABLE chatmessages
    DROP CONSTRAINT chatmessages_chat_id_fk;

DROP TABLE chatmessages;
