BEGIN TRANSACTION;

CREATE TABLE IF NOT EXISTS withdrawals (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    order_num VARCHAR(32) NOT NULL,
    sum NUMERIC(10, 2) NOT NULL,
    processed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

COMMENT ON COLUMN withdrawals.order_num IS 'The order number being placed by the user, it may not be in the DB.';
CREATE INDEX idx_withdrawals_user_id ON withdrawals(user_id);

COMMIT;
