CREATE TABLE conferences
(
    id              UUID PRIMARY KEY,
    title           VARCHAR(100)  NOT NULL,
    description     VARCHAR(1000) NOT NULL,
    speaker_name    VARCHAR(100)  NOT NULL,
    speaker_title   VARCHAR(100)  NOT NULL,
    target_audience VARCHAR(255)  NOT NULL,
    prerequisites   VARCHAR(255),
    seats           INT           NOT NULL,
    starts_at       TIMESTAMP     NOT NULL,
    ends_at         TIMESTAMP     NOT NULL,
    host_id         UUID          NOT NULL REFERENCES users (id),
    status          VARCHAR(50)   NOT NULL DEFAULT 'pending'
        CHECK ( status IN ('pending', 'approved', 'rejected') ),
    created_at      TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at      TIMESTAMP
);

CREATE INDEX conferences_title_idx ON conferences USING gist (title gist_trgm_ops);
CREATE INDEX conferences_status_idx ON conferences (status);
CREATE INDEX conferences_host_id_idx ON conferences (host_id);
CREATE INDEX conferences_starts_at_idx ON conferences (starts_at);
CREATE INDEX conferences_ends_at_idx ON conferences (ends_at);
CREATE INDEX conferences_deleted_at_idx ON conferences (deleted_at);
