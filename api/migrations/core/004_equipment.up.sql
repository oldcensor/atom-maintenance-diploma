CREATE TYPE equipment_status AS ENUM ('active', 'inactive', 'under_maintenance', 'decommissioned');

CREATE TABLE equipment (
    id                  SERIAL           PRIMARY KEY,
    name                VARCHAR(255)     NOT NULL,
    description         TEXT,
    serial_number       VARCHAR(255)     NOT NULL,
    equipment_type_id   INT              NOT NULL REFERENCES equipment_type(id) ON DELETE RESTRICT,
    department_id       INT              REFERENCES department(id) ON DELETE SET NULL,
    responsible_id      INT              REFERENCES employee(id) ON DELETE SET NULL,
    status              equipment_status NOT NULL DEFAULT 'active',
    created_at          TIMESTAMP        NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMP        NOT NULL DEFAULT NOW(),

    CONSTRAINT equipment_serial_uq UNIQUE (serial_number)
);
