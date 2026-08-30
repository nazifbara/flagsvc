-- +goose Up
CREATE TABLE users (
    user_id UUID PRIMARY KEY,
    CONSTRAINT fk_identities FOREIGN KEY (user_id) REFERENCES identities(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL UNIQUE,
    firstName VARCHAR(100) NOT NULL,
    lastName VARCHAR(100) NOT NULL,
    stripeConnectedAccountId VARCHAR(255),
    timezone VARCHAR(50) NOT NULL,
    heartbeat TIMESTAMP,
    lastSignInAt TIMESTAMP,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- +goose Down
DROP TABLE users;