CREATE TABLE exchanges (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    from_currency CHAR(3) NOT NULL,
    to_currency CHAR(3) NOT NULL,
    amount_from BIGINT NOT NULL,
    amount_to BIGINT NOT NULL,
    rate BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);