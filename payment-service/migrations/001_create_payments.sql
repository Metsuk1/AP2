CREATE TABLE IF NOT EXISTS payments (
    id            VARCHAR(36) PRIMARY KEY,
    order_id      VARCHAR(36) NOT NULL,
    transaction_id VARCHAR(36) NOT NULL UNIQUE,
    amount        BIGINT NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'Authorized',
    idempotency_key VARCHAR(64) UNIQUE
);
