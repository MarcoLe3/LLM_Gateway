CREATE TABLE IF NOT EXISTS jobs (
    id                  UUID PRIMARY key,
    caller_id           UUID NOT NULL REFERENCES callers(id),
    idempotency_key     TEXT NOT NULL,
    state               TEXT NOT NULL DEFAULT 'queued'
                        CHECK(state IN ('queued', 'running','succeeded', 'failed', 'dead'))
    attempts            INT NOT NULL DEFAULT 0
    provider            TEXT,
    model               TEXT,
    request_hash        TEXT NOT NULL,
    request_payload     JSONB NOT NULL,
    response_payload    JSONB
    input_tokens        INT,
    output_tokens       INT,
    last_error          TEXT,
    last_error_at       TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    started_at          TIMESTAMPTZ,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    finished_at         TIMESTAMPTZ
    UNIQUE (caller_id, idempotency_key)
)

CREATE TABLE IF NOT EXISTS callers (
    id              UUID PRIMARY key,
    name            TEXT NOT NULL,
    rate_limit_rpm  INT NOT NULL DEFAULT 60,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
)

CREATE TABLE IF NOT EXISTS api_keys (
    id              UUID PRIMARY key,
    caller_id       UUID NOT NULL REFERENCES callers(id)
    api_key_hash    TEXT NOT NULL UNIQUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
    revoked_at      TIMESTAMPTZ
)

CREATE TABLE IF NOT EXISTS usage_events (
    id          UUID PRIMARY key,
    caller_id   UUID NOT NULL REFERENCES callers(id)
    job_id      UUID NOT NULL REFERENCES jobs(id)
    provider    TEXT NOT NULL
)