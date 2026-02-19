CREATE TABLE IF NOT EXISTS measurements (
    id BIGSERIAL PRIMARY KEY,
    sensor_id VARCHAR(50) NOT NULL,
    value DOUBLE PRECISION NOT NULL,
    metadata JSONB, -- Додаткове поле для гнучкості (опціонально)
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


CREATE INDEX IF NOT EXISTS idx_sensor_id ON measurements(sensor_id);


CREATE INDEX IF NOT EXISTS idx_created_at ON measurements(created_at DESC);


CREATE INDEX IF NOT EXISTS idx_sensor_time ON measurements(sensor_id, created_at DESC);