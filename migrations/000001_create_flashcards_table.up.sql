CREATE TABLE IF NOT EXISTS flashcards (
    id BIGSERIAL PRIMARY KEY,
    section TEXT NULL,
    section_type TEXT NULL,
    source_file TEXT NULL,
    text TEXT NULL,
    question TEXT NOT NULL,
    flashcard_type TEXT NOT NULL,
    flashcard_content JSONB NOT NULL,
    categories text[] NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW()
);