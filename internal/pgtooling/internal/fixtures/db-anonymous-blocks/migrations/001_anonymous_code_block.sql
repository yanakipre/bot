---- tern: disable-tx ----

-- This anonymous code block backfills "default" branches column in batches.

DO $$DECLARE
   batch_size CONSTANT INTEGER := 2;
   batch TEXT[];
BEGIN
    LOOP
       SELECT array_agg(id)
         INTO batch
         FROM branches
        WHERE "default" != "primary"
        LIMIT batch_size;

       IF batch IS NULL OR array_length(batch, 1) = 0 THEN
           EXIT;
       END IF;

       UPDATE branches
          SET "default" = "primary"
        WHERE id = ANY(batch);
       -- COMMIT automatically starts a new transaction afterwards
       -- Ref: https://www.postgresql.org/docs/current/plpgsql-transactions.html
       COMMIT;
    END LOOP;
END$$;
