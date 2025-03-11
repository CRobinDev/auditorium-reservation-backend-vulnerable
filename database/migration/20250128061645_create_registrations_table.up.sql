CREATE TABLE registrations
(
    user_id       UUID REFERENCES users (id) ON DELETE CASCADE,
    conference_id UUID REFERENCES conferences (id) ON DELETE CASCADE,
    created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, conference_id)
);
