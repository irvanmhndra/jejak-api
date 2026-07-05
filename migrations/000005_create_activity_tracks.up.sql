CREATE TABLE activity_tracks (
    activity_id UUID PRIMARY KEY REFERENCES activities(id) ON DELETE CASCADE,
    points      JSONB NOT NULL,
    point_count INTEGER NOT NULL DEFAULT 0
);
