CREATE TABLE monitors (
    id UUID PRIMARY KEY,
    url TEXT NOT NULL UNIQUE,
    interval BIGINT NOT NULL,
    timeout BIGINT NOT NULL
);