CREATE TABLE personal_records (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    activity_type VARCHAR(16) NOT NULL,
    record_type   VARCHAR(32) NOT NULL,
    value         DOUBLE PRECISION NOT NULL,
    activity_id   UUID NOT NULL REFERENCES activities(id) ON DELETE CASCADE,
    achieved_at   TIMESTAMPTZ NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, activity_type, record_type)
);
