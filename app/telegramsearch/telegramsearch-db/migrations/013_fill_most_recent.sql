---- tern: disable-tx ----

DO $$DECLARE
    rec RECORD;
BEGIN
    LOOP
        CREATE temp table my_records
        AS SELECT ct.thread_id, ct.body
        FROM chatthreads ct WHERE
            ct.most_recent_message_at = '2011-05-19T09:45:17' LIMIT 100;

        IF (SELECT COUNT(*) FROM my_records) = 0 THEN
            EXIT;
        END IF;

        FOR rec IN SELECT * FROM my_records LOOP
            UPDATE chatthreads
            SET most_recent_message_at = to_timestamp(CAST(
                (rec.body->>(jsonb_array_length(rec.body) - 1))::jsonb->>'date_unixtime'
                AS int)) AT TIME ZONE 'UTC'
            WHERE thread_id = rec.thread_id;
        end loop;

        -- COMMIT automatically starts a new transaction afterwards
        -- Ref: https://www.postgresql.org/docs/current/plpgsql-transactions.html
        COMMIT;

        DROP TABLE my_records;
    END LOOP;
END$$;

