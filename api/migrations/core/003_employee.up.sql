CREATE TYPE employee_role AS ENUM ('technician', 'engineer', 'manager', 'admin');

CREATE TABLE employee (
    id              SERIAL        PRIMARY KEY,
    email           VARCHAR(255)  NOT NULL,
    password_hash   VARCHAR(255)  NOT NULL,
    full_name       VARCHAR(255)  NOT NULL,
    role            employee_role NOT NULL DEFAULT 'technician',
    department_id   INT           REFERENCES department(id) ON DELETE SET NULL,
    failed_attempts INT           NOT NULL DEFAULT 0,
    created_at      TIMESTAMP     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP     NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMP
);

CREATE UNIQUE INDEX employee_email_active_uq ON employee (email) WHERE deleted_at IS NULL;
