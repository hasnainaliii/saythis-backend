CREATE TABLE user_daily_stats (
    id                 UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id            UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date               DATE         NOT NULL,

    mood               VARCHAR(20),
    sleep_hours        NUMERIC(3,1),
    journal_entry      TEXT,
    stress_level       SMALLINT,
    mindful_hours      NUMERIC(3,1),

    stutter_score      NUMERIC(5,1),
    stutter_count      INTEGER,
    repetition_count   INTEGER,
    filler_count       INTEGER,
    total_words        INTEGER,
    stutter_transcript TEXT,

    created_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_user_daily_stats_user_date UNIQUE (user_id, date),
    CONSTRAINT user_daily_stats_mood_check
        CHECK (mood IS NULL OR mood IN ('Dizzy', 'Sad', 'Neutral', 'Happy', 'Great')),
    CONSTRAINT user_daily_stats_sleep_hours_check
        CHECK (sleep_hours IS NULL OR (sleep_hours >= 0 AND sleep_hours <= 12 AND sleep_hours * 2 = ROUND(sleep_hours * 2))),
    CONSTRAINT user_daily_stats_stress_level_check
        CHECK (stress_level IS NULL OR (stress_level >= 1 AND stress_level <= 5)),
    CONSTRAINT user_daily_stats_mindful_hours_check
        CHECK (mindful_hours IS NULL OR (mindful_hours >= 0 AND mindful_hours <= 8 AND mindful_hours * 2 = ROUND(mindful_hours * 2))),
    CONSTRAINT user_daily_stats_stutter_score_check
        CHECK (stutter_score IS NULL OR (stutter_score >= 0 AND stutter_score <= 100)),
    CONSTRAINT user_daily_stats_stutter_count_check
        CHECK (stutter_count IS NULL OR stutter_count >= 0),
    CONSTRAINT user_daily_stats_repetition_count_check
        CHECK (repetition_count IS NULL OR repetition_count >= 0),
    CONSTRAINT user_daily_stats_filler_count_check
        CHECK (filler_count IS NULL OR filler_count >= 0),
    CONSTRAINT user_daily_stats_total_words_check
        CHECK (total_words IS NULL OR total_words >= 0),
    CONSTRAINT user_daily_stats_journal_entry_length_check
        CHECK (journal_entry IS NULL OR char_length(journal_entry) <= 10000),
    CONSTRAINT user_daily_stats_stutter_transcript_length_check
        CHECK (stutter_transcript IS NULL OR char_length(stutter_transcript) <= 50000)
);

CREATE INDEX idx_user_daily_stats_user_date ON user_daily_stats (user_id, date DESC);
CREATE INDEX idx_user_daily_stats_user_journal_date
    ON user_daily_stats (user_id, date DESC)
    WHERE journal_entry IS NOT NULL;
CREATE INDEX idx_user_daily_stats_user_stutter_date
    ON user_daily_stats (user_id, date DESC)
    WHERE stutter_score IS NOT NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_trigger
        WHERE tgname = 'user_daily_stats_updated_at'
    ) THEN
        CREATE TRIGGER user_daily_stats_updated_at
            BEFORE UPDATE ON user_daily_stats
            FOR EACH ROW EXECUTE FUNCTION update_updated_at();
    END IF;
END;
$$;
