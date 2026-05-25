CREATE TABLE exercise_progress (
    id           UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    chapter_id   VARCHAR(50)  NOT NULL,
    exercise_id  VARCHAR(50)  NOT NULL,
    completed    BOOLEAN      NOT NULL DEFAULT TRUE,
    rating       INT          NOT NULL CHECK (rating >= 1 AND rating <= 5),
    remarks      TEXT         NOT NULL DEFAULT '',
    completed_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_user_exercise UNIQUE (user_id, exercise_id)
);

CREATE INDEX idx_exercise_progress_user_id ON exercise_progress (user_id);
