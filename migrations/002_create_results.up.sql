CREATE TABLE results (
    id BIGSERIAL PRIMARY KEY,
    monitor_id UUID NOT NULL REFERENCES monitors(id),
    url TEXT NOT NULL,
    response_code INTEGER,
    latency BIGINT,
    success BOOLEAN NOT NULL,
    checked_at TIMESTAMPTZ NOT NULL,
    failure_reason TEXT
);