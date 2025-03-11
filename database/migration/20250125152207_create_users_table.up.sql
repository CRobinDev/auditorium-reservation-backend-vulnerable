CREATE TABLE users (
	id UUID PRIMARY KEY,
	name VARCHAR(100) NOT NULL,
	email VARCHAR(320) NOT NULL,
	password_hash CHAR(60) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'user' CHECK ( role IN ('admin', 'event_coordinator', 'user') ),
    bio VARCHAR(500),
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	deleted_at TIMESTAMP
);

CREATE UNIQUE INDEX users_email_key ON users (email) WHERE deleted_at IS NULL;
CREATE INDEX users_deleted_at_idx ON users (deleted_at);
