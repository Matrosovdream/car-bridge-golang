CREATE TABLE IF NOT EXISTS saferweb_company (
    id         BIGSERIAL PRIMARY KEY,
    dot_number TEXT        NOT NULL UNIQUE,
    legal_name TEXT        NOT NULL DEFAULT '',
    dba_name   TEXT        NOT NULL DEFAULT '',
    raw_json   JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
