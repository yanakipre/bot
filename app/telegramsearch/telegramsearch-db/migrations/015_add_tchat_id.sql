ALTER TABLE chats ADD COLUMN telegram_chat_id text NOT NULL DEFAULT 'empty';

---- tern: disable-tx ----
