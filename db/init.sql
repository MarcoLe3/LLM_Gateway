CREATE TABLE IF NOT EXISTS jobs (
    id                  UUID PRIMARY key,
    caller_id           TEXT NOT NULL,
    idempotency_key     TEXT NOT NULL,
    state               TEXT NOT NULL
    provider            TEXT,
    model               TEXT,
    request_payload     JSONB,
    response_payload    JSONB
    input_tokens        INT,
    output_tokens       INT,
    last_error          TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    started_at          TIMESTAMPTZ,
    updated_at          TIMESTAMPTZ,
    finished_at         TIMESTAMPTZ
)

CREATE TABLE IF NOT EXISTS callers (
    id              UUID PRIMARY key,
    name            TEXT NOT NULL,
    api_key_hash    TEXT NOT NULL UNIQUE,
    rate_limit_rpm  INT NOT NULL DEFAULT 60,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
)