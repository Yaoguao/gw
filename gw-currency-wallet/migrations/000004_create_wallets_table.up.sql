CREATE TABLE wallets (
     id UUID PRIMARY KEY,
     user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
     currency CHAR(3) NOT NULL,
     balance BIGINT NOT NULL DEFAULT 0,
     UNIQUE(user_id, currency)
);