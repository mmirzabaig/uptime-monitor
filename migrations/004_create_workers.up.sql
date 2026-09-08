CREATE TABLE workers (
    id UUID PRIMARY KEY,
    address TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    last_heartbeat TIMESTAMPTZ,
    lease_expires_at TIMESTAMPTZ
);