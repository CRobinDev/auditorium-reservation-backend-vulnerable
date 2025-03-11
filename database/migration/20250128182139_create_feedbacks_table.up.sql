CREATE TABLE feedbacks
(
    id            UUID PRIMARY KEY,
    user_id       UUID          NOT NULL REFERENCES users (id),
    conference_id UUID          NOT NULL REFERENCES conferences (id) ON DELETE CASCADE,
    comment       VARCHAR(1000) NOT NULL,
    created_at    TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at    TIMESTAMP
);

CREATE INDEX feedbacks_conference_id_idx ON feedbacks (conference_id);
CREATE INDEX feedbacks_deleted_at_idx ON feedbacks (deleted_at);
