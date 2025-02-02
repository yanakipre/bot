---- tern: disable-tx ----

DO $$DECLARE
    batch BIGINT[];
BEGIN
    LOOP
        SELECT array_agg(em.embedding_id)
        INTO batch
        FROM embeddings em JOIN chatthreads thr ON (em.thread_id = thr.thread_id) WHERE
            em.chat_id != thr.chat_id LIMIT 10;

        IF batch IS NULL OR array_length(batch, 1) = 0 THEN
            EXIT;
        END IF;

        UPDATE embeddings em
        SET chat_id = (SELECT chat_id FROM chatthreads WHERE em.thread_id = thread_id)
        WHERE em.embedding_id = ANY(batch);

        -- COMMIT automatically starts a new transaction afterwards
        -- Ref: https://www.postgresql.org/docs/current/plpgsql-transactions.html
        COMMIT;
    END LOOP;
END$$;

