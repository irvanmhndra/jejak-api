CREATE TABLE activities (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type              VARCHAR(16) NOT NULL,
    title             VARCHAR(200) NOT NULL,
    notes             TEXT,
    source            VARCHAR(16) NOT NULL DEFAULT 'gps',
    started_at        TIMESTAMPTZ NOT NULL,
    elapsed_seconds   INTEGER NOT NULL DEFAULT 0,
    moving_seconds    INTEGER NOT NULL DEFAULT 0,
    distance_meters   DOUBLE PRECISION NOT NULL DEFAULT 0,
    avg_speed_mps     DOUBLE PRECISION NOT NULL DEFAULT 0,
    max_speed_mps     DOUBLE PRECISION NOT NULL DEFAULT 0,
    elevation_gain_m  DOUBLE PRECISION NOT NULL DEFAULT 0,
    elevation_loss_m  DOUBLE PRECISION NOT NULL DEFAULT 0,
    calories          INTEGER,
    polyline          TEXT NOT NULL DEFAULT '',
    point_count       INTEGER NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_activities_user_started ON activities(user_id, started_at DESC);
CREATE INDEX idx_activities_user_type    ON activities(user_id, type);
