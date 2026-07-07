CREATE TABLE goals (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    period        VARCHAR(16) NOT NULL,  -- 'weekly' | 'monthly'
    activity_type VARCHAR(16),           -- NULL = semua tipe
    target_meters DOUBLE PRECISION NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_goals_user ON goals(user_id);
