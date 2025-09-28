CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    description VARCHAR(500) NOT NULL,
    owner_id UUID NOT NULL,
    FOREIGN KEY (owner_id) REFERENCES users(id),
    cpu_limit INT NOT NULL,
    ram_limit INT NOT NULL,
    storage_limit INT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_projects_owner_id ON projects(owner_id);