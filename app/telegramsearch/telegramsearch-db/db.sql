CREATE EXTENSION IF NOT EXISTS vector WITH SCHEMA public;

CREATE TABLE public.chats (
    chat_id text NOT NULL,
    telegram_chat_id text DEFAULT 'empty'::text NOT NULL
);

CREATE TABLE public.chatthreads (
    thread_id bigint NOT NULL,
    chat_id text NOT NULL,
    body jsonb NOT NULL,
    most_recent_message_at timestamp with time zone DEFAULT '2011-05-19 09:45:17+00'::timestamp with time zone
);

CREATE SEQUENCE public.chatthreads_thread_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.chatthreads_thread_id_seq OWNED BY public.chatthreads.thread_id;

CREATE TABLE public.embeddings (
    thread_id bigint NOT NULL,
    chat_id text NOT NULL,
    message text NOT NULL,
    embedding public.vector(2000),
    embedding_id bigint NOT NULL
);

CREATE SEQUENCE public.embeddings_embedding_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.embeddings_embedding_id_seq OWNED BY public.embeddings.embedding_id;

CREATE SEQUENCE public.embeddings_thread_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.embeddings_thread_id_seq OWNED BY public.embeddings.thread_id;

CREATE TABLE public.schema_version (
    version integer NOT NULL
);

ALTER TABLE ONLY public.chatthreads ALTER COLUMN thread_id SET DEFAULT nextval('public.chatthreads_thread_id_seq'::regclass);

ALTER TABLE ONLY public.embeddings ALTER COLUMN thread_id SET DEFAULT nextval('public.embeddings_thread_id_seq'::regclass);

ALTER TABLE ONLY public.embeddings ALTER COLUMN embedding_id SET DEFAULT nextval('public.embeddings_embedding_id_seq'::regclass);

ALTER TABLE ONLY public.chats
    ADD CONSTRAINT chats_pkey PRIMARY KEY (chat_id);

ALTER TABLE ONLY public.chatthreads
    ADD CONSTRAINT chatthreads_pkey PRIMARY KEY (thread_id);

ALTER TABLE ONLY public.embeddings
    ADD CONSTRAINT embeddings_pkey PRIMARY KEY (embedding_id);

CREATE INDEX chatthreads_chat_id_idx ON public.chatthreads USING hash (chat_id);

CREATE INDEX embeddings_2000_idx ON public.embeddings USING hnsw (embedding public.vector_l2_ops);

CREATE INDEX embeddings_chat_id_idx ON public.embeddings USING hash (chat_id);

CREATE UNIQUE INDEX embeddings_chatthread_id_idx ON public.embeddings USING btree (thread_id);

CREATE INDEX embeddings_most_recent_message_at_idx ON public.chatthreads USING btree (most_recent_message_at);

ALTER TABLE ONLY public.chatthreads
    ADD CONSTRAINT chatthreads_chat_id_fk FOREIGN KEY (chat_id) REFERENCES public.chats(chat_id) ON DELETE CASCADE;

ALTER TABLE ONLY public.embeddings
    ADD CONSTRAINT embeddings_chat_id_fk FOREIGN KEY (chat_id) REFERENCES public.chats(chat_id) ON DELETE CASCADE;

ALTER TABLE ONLY public.embeddings
    ADD CONSTRAINT embeddings_chatthread_id_fk FOREIGN KEY (thread_id) REFERENCES public.chatthreads(thread_id) ON DELETE CASCADE;