BEGIN TRANSACTION;

ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS attempts SMALLINT DEFAULT 0;

COMMENT ON COLUMN orders.attempts IS   
    'The number of unsuccessful order processing attempts';

COMMIT;
