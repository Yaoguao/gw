CREATE TABLE IF NOT EXISTS exchange_rates (
    from_currency VARCHAR(3) NOT NULL,
    to_currency VARCHAR(3) NOT NULL,
    rate DOUBLE PRECISION NOT NULL,
    updated_at TIMESTAMP DEFAULT now(),
    PRIMARY KEY (from_currency, to_currency)
);