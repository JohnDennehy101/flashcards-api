CREATE TABLE IF NOT EXISTS flashcards (
    id BIGSERIAL PRIMARY KEY,
    section TEXT NOT NULL DEFAULT '',
    section_type TEXT NOT NULL DEFAULT '',
    source_file TEXT NOT NULL DEFAULT '',
    text TEXT NOT NULL DEFAULT '',
    question TEXT NOT NULL,
    flashcard_type TEXT NOT NULL,
    flashcard_content JSONB NOT NULL,
    categories text[] NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW()
    );