CREATE TABLE models (
    id SERIAL PRIMARY KEY,
    field1 TEXT NOT NULL,
    field2 INT NOT NULL,
    created_by INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    modified_by INT,
    modified_at TIMESTAMP,
    deleted_by INT,
    deleted_at TIMESTAMP
);