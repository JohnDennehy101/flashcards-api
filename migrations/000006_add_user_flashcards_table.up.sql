CREATE TABLE IF NOT EXISTS user_flashcards (
                                               user_id bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    flashcard_id bigint NOT NULL REFERENCES flashcards(id) ON DELETE CASCADE,
    correct_count integer NOT NULL DEFAULT 0,
    status text NOT NULL DEFAULT 'not_started',
    last_reviewed_at timestamp(0) with time zone DEFAULT NOW(),
    PRIMARY KEY (user_id, flashcard_id)
    );

CREATE INDEX idx_user_flashcards_status ON user_flashcards(user_id, status);