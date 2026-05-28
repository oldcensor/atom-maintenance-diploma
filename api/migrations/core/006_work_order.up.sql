CREATE TYPE work_order_status AS ENUM ('open', 'in_progress', 'completed', 'cancelled');
CREATE TYPE work_order_type  AS ENUM ('emergency', 'corrective', 'planned');

CREATE TABLE work_order (
    id           SERIAL            PRIMARY KEY,
    schedule_id  INT               REFERENCES maintenance_schedule(id) ON DELETE SET NULL,
    equipment_id INT               NOT NULL REFERENCES equipment(id) ON DELETE RESTRICT,
    title        VARCHAR(255)      NOT NULL,
    description  TEXT,
    assigned_to  INT               REFERENCES employee(id) ON DELETE SET NULL,
    created_by   INT               REFERENCES employee(id) ON DELETE SET NULL,
    status       work_order_status NOT NULL DEFAULT 'open',
    work_type    work_order_type   NOT NULL DEFAULT 'corrective',
    created_at   TIMESTAMP         NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMP         NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP,
    CONSTRAINT work_order_planned_requires_schedule CHECK (
        (work_type = 'planned'    AND schedule_id IS NOT NULL) OR
        (work_type IN ('corrective', 'emergency'))
    )
);

CREATE TABLE work_order_status_log (
    id            SERIAL            PRIMARY KEY,
    work_order_id INT               NOT NULL REFERENCES work_order(id) ON DELETE CASCADE,
    from_status   work_order_status NOT NULL,
    to_status     work_order_status NOT NULL,
    changed_by    INT               REFERENCES employee(id) ON DELETE SET NULL,
    comment       TEXT,
    created_at    TIMESTAMP         NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_wo_status_log_wo_id ON work_order_status_log(work_order_id, created_at);

CREATE TABLE work_order_comment (
    id            SERIAL    PRIMARY KEY,
    work_order_id INT       NOT NULL REFERENCES work_order(id) ON DELETE CASCADE,
    author_id     INT       NOT NULL REFERENCES employee(id) ON DELETE CASCADE,
    text          TEXT      NOT NULL,
    created_at    TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_wo_comment_wo_id ON work_order_comment(work_order_id, created_at);

CREATE TABLE work_order_checklist_item (
    id            SERIAL    PRIMARY KEY,
    work_order_id INT       NOT NULL REFERENCES work_order(id) ON DELETE CASCADE,
    text          TEXT      NOT NULL,
    checked       BOOLEAN   NOT NULL DEFAULT FALSE,
    checked_by    INT       REFERENCES employee(id) ON DELETE SET NULL,
    checked_at    TIMESTAMP,
    sort_order    INT       NOT NULL DEFAULT 0,
    created_at    TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_wo_checklist_wo_id ON work_order_checklist_item(work_order_id, sort_order);