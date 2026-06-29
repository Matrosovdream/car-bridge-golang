CREATE TABLE IF NOT EXISTS request_log (
    id          BIGSERIAL PRIMARY KEY,
    provider    TEXT        NOT NULL,
    operation   TEXT        NOT NULL,
    request_ref TEXT,
    status_code INTEGER,
    success     BOOLEAN     NOT NULL DEFAULT FALSE,
    latency_ms  INTEGER,
    error       TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_request_log_provider ON request_log (provider, created_at DESC);
