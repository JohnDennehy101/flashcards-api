CREATE INDEX IF NOT EXISTS flashcards_section_idx ON flashcards USING GIN (to_tsvector('simple', section));
CREATE INDEX IF NOT EXISTS flashcards_section_type_idx
    ON flashcards (section_type);
CREATE INDEX IF NOT EXISTS flashcards_source_file_idx
    ON flashcards (source_file);
CREATE INDEX IF NOT EXISTS flashcards_type_idx
    ON flashcards (flashcard_type);
CREATE INDEX IF NOT EXISTS flashcards_categories_idx ON flashcards USING GIN (categories);