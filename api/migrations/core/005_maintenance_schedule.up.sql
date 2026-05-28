CREATE TYPE schedule_status AS ENUM ('scheduled', 'in_progress', 'completed', 'cancelled');

CREATE TABLE maintenance_schedule (
    id               SERIAL          PRIMARY KEY,
    equipment_id     INT             NOT NULL REFERENCES equipment(id) ON DELETE CASCADE,
    scheduled_at     TIMESTAMP       NOT NULL,
    description      TEXT,
    assigned_to      INT             REFERENCES employee(id) ON DELETE SET NULL,
    status           schedule_status NOT NULL DEFAULT 'scheduled',
    interval_unit    VARCHAR(20),
    interval_value   INT,
    last_meter_value FLOAT,
    next_due_at      DATE,
    created_at       TIMESTAMP       NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMP       NOT NULL DEFAULT NOW()
);
