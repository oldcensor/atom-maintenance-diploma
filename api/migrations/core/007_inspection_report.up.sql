CREATE TABLE inspection_report (
    id              SERIAL    PRIMARY KEY,
    work_order_id   INT       NOT NULL REFERENCES work_order(id) ON DELETE RESTRICT,
    inspector_id    INT       NOT NULL REFERENCES employee(id) ON DELETE RESTRICT,
    findings        TEXT      NOT NULL,
    recommendations TEXT,
    created_at      TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Один протокол на наряд (отношение 1:1)
CREATE UNIQUE INDEX idx_inspection_report_work_order ON inspection_report(work_order_id);
