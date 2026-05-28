CREATE TABLE department (
    id          SERIAL       PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    description TEXT,
    created_at  TIMESTAMP    NOT NULL DEFAULT NOW(),

    CONSTRAINT department_name_uq UNIQUE (name)
);
